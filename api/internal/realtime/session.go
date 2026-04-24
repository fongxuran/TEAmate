package realtime

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"nhooyr.io/websocket"

	"teammate/internal/analysis"
	"teammate/internal/model"
)

type Session struct {
	id string

	mu sync.Mutex

	agendaText     string
	agendaItems    []model.AgendaItem
	transcriptText string
	config         analysis.Config

	feedbackOverrides map[string]bool
	result            analysis.Result
	lastSegmentIDs    map[string]struct{}
	alertedSegmentIDs map[string]struct{}

	conns map[*websocket.Conn]struct{}
}

func newSession(id string) *Session {
	return &Session{
		id:                id,
		config:            analysis.DefaultConfig(),
		feedbackOverrides: make(map[string]bool),
		lastSegmentIDs:    make(map[string]struct{}),
		alertedSegmentIDs: make(map[string]struct{}),
		conns:             make(map[*websocket.Conn]struct{}),
	}
}

func (s *Session) addConn(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[conn] = struct{}{}
}

func (s *Session) removeConn(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, conn)
}

func (s *Session) broadcast(evt serverEvent) {
	s.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(s.conns))
	for c := range s.conns {
		conns = append(conns, c)
	}
	s.mu.Unlock()

	// Fire-and-forget writes with per-conn timeout.
	for _, c := range conns {
		cc := c
		go func() {
			_ = writeJSON(context.Background(), cc, evt)
		}()
	}
}

func (s *Session) setAgenda(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agendaText = text
	s.agendaItems = analysis.ParseAgendaText(text)
}

func (s *Session) setConfig(cfg analysis.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
}

func (s *Session) applyRealtimeMessage(msg model.RealtimeMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg.TextDelta != nil {
		s.transcriptText += *msg.TextDelta
		return
	}
	if msg.Text != nil {
		s.transcriptText = *msg.Text
	}
}

func (s *Session) applyDriftFeedback(segmentID string, isDrift bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feedbackOverrides[segmentID] = isDrift
}

func (s *Session) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agendaText = ""
	s.agendaItems = nil
	s.transcriptText = ""
	s.feedbackOverrides = make(map[string]bool)
	s.lastSegmentIDs = make(map[string]struct{})
	s.alertedSegmentIDs = make(map[string]struct{})
	s.result = analysis.Result{}
}

// Reset clears all state for the session.
func (s *Session) Reset() {
	s.reset()
}

// Recompute runs deterministic analysis for the current state and returns the latest result.
func (s *Session) Recompute() analysis.Result {
	_, _ = s.recompute(false)
	return s.analysisPayload()
}

type driftAlert struct {
	SegmentID       string  `json:"segment_id"`
	BestAgendaTitle *string `json:"best_agenda_title,omitempty"`
	BestScore       float64 `json:"best_score"`
	TextPreview     string  `json:"text_preview"`
}

// recompute updates analysis.Result. If emitAlerts is true, returns drift alerts for newly drift segments.
// Returns (updated, alerts).
func (s *Session) recompute(emitAlerts bool) (bool, []driftAlert) {
	s.mu.Lock()
	meeting := model.MeetingInput{
		SchemaVersion: "v1",
		Agenda:        append([]model.AgendaItem(nil), s.agendaItems...),
		Transcript:    analysis.ParseTranscriptText(s.transcriptText),
	}
	cfg := s.config
	feedback := make(map[string]bool, len(s.feedbackOverrides))
	for k, v := range s.feedbackOverrides {
		feedback[k] = v
	}
	prevIDs := s.lastSegmentIDs
	newLast := make(map[string]struct{})
	s.mu.Unlock()

	res := analysis.Analyze(meeting, cfg, feedback)

	alerts := make([]driftAlert, 0)
	for _, seg := range res.Segments {
		newLast[seg.Segment.SegmentID] = struct{}{}
		if !emitAlerts {
			continue
		}
		if !seg.IsDrift {
			continue
		}
		// Alert only once per segment and only when it appears new.
		if _, existed := prevIDs[seg.Segment.SegmentID]; !existed {
			alerts = append(alerts, driftAlert{
				SegmentID:       seg.Segment.SegmentID,
				BestAgendaTitle: seg.BestAgendaTitle,
				BestScore:       seg.BestScore,
				TextPreview:     preview(seg.Segment.Text, 180),
			})
		}
	}

	s.mu.Lock()
	updated := true
	s.result = res
	s.lastSegmentIDs = newLast
	// Track alerted segments.
	for _, a := range alerts {
		s.alertedSegmentIDs[a.SegmentID] = struct{}{}
	}
	s.mu.Unlock()

	return updated, alerts
}

func preview(s string, max int) string {
	ss := strings.TrimSpace(s)
	if len(ss) <= max {
		return ss
	}
	return ss[:max] + "…"
}

func (s *Session) snapshotPayload() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"session_id":      s.id,
		"agenda_text":     s.agendaText,
		"agenda":          append([]model.AgendaItem(nil), s.agendaItems...),
		"transcript_text": s.transcriptText,
		"config":          s.config,
		"analysis":        s.result,
	}
}

func (s *Session) analysisPayload() analysis.Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result
}

func (s *Session) Snapshot() (agendaText string, transcriptText string, agenda []model.AgendaItem, cfg analysis.Config, res analysis.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.agendaText, s.transcriptText, append([]model.AgendaItem(nil), s.agendaItems...), s.config, s.result
}

// MarshalSnapshot is helpful for REST responses.
func (s *Session) MarshalSnapshot() ([]byte, error) {
	agendaText, transcriptText, agenda, cfg, res := s.Snapshot()
	payload := map[string]any{
		"session_id":      s.id,
		"agenda_text":     agendaText,
		"agenda":          agenda,
		"transcript_text": transcriptText,
		"config":          cfg,
		"analysis":        res,
	}
	return json.Marshal(payload)
}
