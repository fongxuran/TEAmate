package analysis

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"teammate/internal/model"
)

// Config controls deterministic MVP scoring behavior.
type Config struct {
	// DriftThreshold marks a segment as drift when bestScore < DriftThreshold.
	DriftThreshold float64 `json:"drift_threshold"`
	// SegmentMaxTokens is passed to model.SegmentTranscript.
	SegmentMaxTokens int `json:"segment_max_tokens"`
	// SegmentMaxChars is passed to model.SegmentTranscript.
	SegmentMaxChars int `json:"segment_max_chars"`
}

func (c Config) withDefaults() Config {
	if c.DriftThreshold <= 0 {
		c.DriftThreshold = 0.08
	}
	if c.SegmentMaxTokens <= 0 {
		c.SegmentMaxTokens = 220
	}
	if c.SegmentMaxChars <= 0 {
		c.SegmentMaxChars = 1800
	}
	return c
}

// DriftSegment is a segment enriched with deterministic drift scoring.
type DriftSegment struct {
	Segment          model.Segment `json:"segment"`
	BestAgendaItemID *string       `json:"best_agenda_item_id,omitempty"`
	BestAgendaTitle  *string       `json:"best_agenda_title,omitempty"`
	BestScore        float64       `json:"best_score"`
	IsDrift          bool          `json:"is_drift"`
	// FeedbackOverride is non-nil when a client has explicitly marked drift / not drift.
	FeedbackOverride *bool `json:"feedback_override,omitempty"`
}

// Result is the end-to-end analysis output consumed by the local UI.
type Result struct {
	SchemaVersion string         `json:"schema_version"`
	GeneratedAt   time.Time      `json:"generated_at"`
	Agenda        []model.AgendaItem `json:"agenda"`
	Transcript    model.Transcript   `json:"transcript"`
	Segments      []DriftSegment `json:"segments"`
	Summary       string         `json:"summary"`
	Decisions     []string       `json:"decisions"`
	ActionItems   []model.ActionItem `json:"action_items"`
	TicketDrafts  []model.TicketDraft `json:"ticket_drafts"`
}

var nonWord = regexp.MustCompile(`[^a-z0-9]+`)

// ParseAgendaText converts a plain multiline agenda textbox into canonical agenda items.
// Each non-empty line becomes one agenda item.
func ParseAgendaText(agendaText string) []model.AgendaItem {
	lines := strings.Split(agendaText, "\n")
	out := make([]model.AgendaItem, 0, len(lines))
	idx := 0
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		line = strings.TrimLeft(line, "-•*\t ")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx++
		id := fmt.Sprintf("a-%02d", idx)
		keywords := extractKeywords(line)
		out = append(out, model.AgendaItem{ID: id, Title: line, Keywords: keywords})
	}
	return out
}

// ParseTranscriptText converts a plain multiline transcript into a turn-based transcript.
// Each non-empty line becomes a turn. If the line looks like "Speaker: text", speaker is set.
func ParseTranscriptText(transcriptText string) model.Transcript {
	lines := strings.Split(transcriptText, "\n")
	turns := make([]model.TranscriptTurn, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		turn := model.TranscriptTurn{Text: line}
		if i := strings.Index(line, ":"); i > 0 && i < 40 {
			spk := strings.TrimSpace(line[:i])
			rest := strings.TrimSpace(line[i+1:])
			if spk != "" && rest != "" {
				turn.Speaker = &spk
				turn.Text = rest
			}
		}
		turns = append(turns, turn)
	}
	return model.Transcript{Turns: turns}
}

// Analyze performs deterministic segmentation + drift scoring + lightweight outcome extraction.
// feedbackOverrides maps segment_id -> is_drift override.
func Analyze(meeting model.MeetingInput, cfg Config, feedbackOverrides map[string]bool) Result {
	cfg = cfg.withDefaults()

	segs := model.SegmentTranscript(meeting.Transcript, model.SegmentOptions{
		MaxTokens:                 cfg.SegmentMaxTokens,
		MaxChars:                  cfg.SegmentMaxChars,
		IncludeSpeakerLabels:      true,
		ComputeSpeakerDistribution: false,
	})

	items := meeting.Agenda
	if items == nil {
		items = []model.AgendaItem{}
	}

	// Drift scoring.
	segments := make([]DriftSegment, 0, len(segs))
	for _, s := range segs {
		bestID, bestTitle, bestScore := bestAgendaMatch(items, s.Text)
		isDrift := false
		if len(items) > 0 {
			isDrift = bestScore < cfg.DriftThreshold
		}

		ds := DriftSegment{
			Segment:   s,
			BestScore: bestScore,
			IsDrift:   isDrift,
		}
		if bestID != "" {
			ds.BestAgendaItemID = &bestID
		}
		if bestTitle != "" {
			ds.BestAgendaTitle = &bestTitle
		}
		if feedbackOverrides != nil {
			if v, ok := feedbackOverrides[s.SegmentID]; ok {
				vv := v
				ds.FeedbackOverride = &vv
				ds.IsDrift = v
			}
		}
		segments = append(segments, ds)
	}

	// Outcomes: simple, deterministic extraction.
	full := strings.TrimSpace(transcriptText(meeting))
	summary := summarizePlaintext(full)
	decisions := extractPrefixed(full, []string{"decision:", "decisions:"})
	actionItems := extractActionItems(meeting, full)
	drafts := model.TicketDraftsFromActionItems(meeting, actionItems)

	return Result{
		SchemaVersion: "v1",
		GeneratedAt:   time.Now().UTC(),
		Agenda:        append([]model.AgendaItem(nil), items...),
		Transcript:    meeting.Transcript,
		Segments:      segments,
		Summary:       summary,
		Decisions:     decisions,
		ActionItems:   actionItems,
		TicketDrafts:  drafts,
	}
}

func extractKeywords(s string) []string {
	set := make(map[string]struct{})
	for _, w := range strings.Fields(strings.ToLower(s)) {
		w = nonWord.ReplaceAllString(w, "")
		if len(w) < 4 {
			continue
		}
		if isStopword(w) {
			continue
		}
		set[w] = struct{}{}
	}
	res := make([]string, 0, len(set))
	for w := range set {
		res = append(res, w)
	}
	sort.Strings(res)
	return res
}

func bestAgendaMatch(items []model.AgendaItem, segmentText string) (id string, title string, score float64) {
	if len(items) == 0 {
		return "", "", 0
	}
	segWords := wordSet(segmentText)
	bestScore := -1.0
	bestIdx := -1
	for i := range items {
		w := wordSet(items[i].Title + " " + strings.Join(items[i].Keywords, " "))
		s := jaccard(segWords, w)
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return "", "", 0
	}
	return items[bestIdx].ID, items[bestIdx].Title, bestScore
}

func wordSet(s string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, w := range strings.Fields(strings.ToLower(s)) {
		w = nonWord.ReplaceAllString(w, "")
		if w == "" || len(w) < 3 {
			continue
		}
		if isStopword(w) {
			continue
		}
		set[w] = struct{}{}
	}
	return set
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	union := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		union[k] = struct{}{}
	}
	for k := range b {
		if _, ok := a[k]; ok {
			inter++
		}
		union[k] = struct{}{}
	}
	return float64(inter) / float64(len(union))
}

func isStopword(w string) bool {
	switch w {
	case "the", "and", "for", "with", "that", "this", "from", "into", "your", "you", "our", "are", "was", "were", "will", "have", "has", "had", "not", "but", "can", "could", "should", "would", "about", "just", "than", "then", "them", "they", "their", "what", "when", "where", "who", "why", "how", "lets", "let", "today", "next", "also", "maybe", "like":
		return true
	default:
		return false
	}
}

func summarizePlaintext(s string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, 3)
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		out = append(out, l)
		if len(out) >= 3 {
			break
		}
	}
	return strings.Join(out, " ")
}

func extractPrefixed(s string, prefixes []string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0)
	for _, l := range lines {
		trim := strings.TrimSpace(l)
		low := strings.ToLower(trim)
		for _, p := range prefixes {
			if strings.HasPrefix(low, p) {
				v := strings.TrimSpace(trim[len(p):])
				if v != "" {
					out = append(out, v)
				}
				break
			}
		}
	}
	return out
}

func extractActionItems(meeting model.MeetingInput, s string) []model.ActionItem {
	items := extractPrefixed(s, []string{"action:", "action item:", "action items:"})
	if len(items) == 0 {
		return nil
	}
	out := make([]model.ActionItem, 0, len(items))
	for i, title := range items {
		id := stableID(meeting, title, i)
		out = append(out, model.ActionItem{
			ActionItemID: id,
			Title:        title,
			Confidence:   0.6,
		})
	}
	return out
}

func stableID(meeting model.MeetingInput, title string, idx int) string {
	meetingHint := ""
	if meeting.Transcript.MeetingID != nil {
		meetingHint = *meeting.Transcript.MeetingID
	} else if meeting.Transcript.MeetingName != nil {
		meetingHint = *meeting.Transcript.MeetingName
	}
	h := sha1.Sum([]byte(fmt.Sprintf("%s|%d|%s", meetingHint, idx, strings.TrimSpace(title))))
	return "ai-" + hex.EncodeToString(h[:8])
}

func transcriptText(m model.MeetingInput) string {
	if len(m.Transcript.Turns) == 0 {
		return ""
	}
	var b strings.Builder
	for i, t := range m.Transcript.Turns {
		if i > 0 {
			b.WriteByte('\n')
		}
		if t.Speaker != nil && strings.TrimSpace(*t.Speaker) != "" {
			b.WriteString(strings.TrimSpace(*t.Speaker))
			b.WriteString(": ")
		}
		b.WriteString(strings.TrimSpace(t.Text))
	}
	return b.String()
}
