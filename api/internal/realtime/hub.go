package realtime

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"

	"teammate/internal/analysis"
	"teammate/internal/model"
)

// Hub manages realtime sessions (in-memory) for the local MVP.
type Hub struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

func NewHub() *Hub {
	return &Hub{sessions: make(map[string]*Session)}
}

// GetSession returns (and lazily creates) an in-memory session.
// An empty id maps to "default".
func (h *Hub) GetSession(id string) *Session {
	if id == "" {
		id = "default"
	}
	return h.getSession(id)
}

func (h *Hub) getSession(id string) *Session {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s, ok := h.sessions[id]; ok {
		return s
	}
	s := newSession(id)
	h.sessions[id] = s
	return s
}

// ServeWS upgrades the connection and joins a session.
// This endpoint is intentionally NOT mounted under /api so browser clients can connect
// without needing custom headers (e.g., Basic Auth).
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = "default"
	}

	// For local dev, allow cross-origin connections from the Next.js dev server.
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusInternalError, "internal error")

	s := h.GetSession(sessionID)
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = "anonymous"
	}

	s.addConn(conn)
	defer s.removeConn(conn)

	// Ensure sync includes a computed analysis payload.
	_ = s.Recompute()

	// Initial sync.
	_ = writeJSON(ctx, conn, serverEvent{Type: "sync", Payload: s.snapshotPayload()})

	for {
		msgType, data, err := conn.Read(ctx)
		if err != nil {
			break
		}
		if msgType != websocket.MessageText {
			continue
		}
		var evt clientEvent
		if err := json.Unmarshal(data, &evt); err != nil {
			continue
		}
		// Apply.
		switch evt.Type {
		case "set_agenda":
			var p setAgendaPayload
			if json.Unmarshal(evt.Payload, &p) == nil {
				s.setAgenda(p.AgendaText)
				s.broadcast(serverEvent{Type: "agenda_updated", Payload: s.snapshotPayload()})
				// Recompute analysis when agenda changes.
				updates, alerts := s.recompute(true)
				if updates {
					s.broadcast(serverEvent{Type: "analysis_updated", Payload: s.analysisPayload()})
				}
				for _, a := range alerts {
					s.broadcast(serverEvent{Type: "drift_alert", Payload: a})
				}
			}
		case "set_config":
			var p analysis.Config
			if json.Unmarshal(evt.Payload, &p) == nil {
				s.setConfig(p)
				updates, alerts := s.recompute(false)
				if updates {
					s.broadcast(serverEvent{Type: "analysis_updated", Payload: s.analysisPayload()})
				}
				for _, a := range alerts {
					s.broadcast(serverEvent{Type: "drift_alert", Payload: a})
				}
			}
		case "realtime_message":
			var p analysisRealtimeMessage
			if json.Unmarshal(evt.Payload, &p) == nil {
				p.Message.ClientID = clientID
				if p.Message.Timestamp.IsZero() {
					p.Message.Timestamp = time.Now().UTC()
				}
				s.applyRealtimeMessage(p.Message)
				s.broadcast(serverEvent{Type: "transcript_updated", Payload: p.Message})
				updates, alerts := s.recompute(true)
				if updates {
					s.broadcast(serverEvent{Type: "analysis_updated", Payload: s.analysisPayload()})
				}
				for _, a := range alerts {
					s.broadcast(serverEvent{Type: "drift_alert", Payload: a})
				}
			}
		case "drift_feedback":
			var p driftFeedbackPayload
			if json.Unmarshal(evt.Payload, &p) == nil && p.SegmentID != "" {
				s.applyDriftFeedback(p.SegmentID, p.IsDrift)
				s.broadcast(serverEvent{Type: "drift_feedback_applied", Payload: p})
				updates, _ := s.recompute(false)
				if updates {
					s.broadcast(serverEvent{Type: "analysis_updated", Payload: s.analysisPayload()})
				}
			}
		case "reset":
			s.reset()
			s.broadcast(serverEvent{Type: "reset_applied", Payload: s.snapshotPayload()})
		default:
			// ignore
		}
	}

	_ = conn.Close(websocket.StatusNormalClosure, "")
}

type clientEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type serverEvent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

type setAgendaPayload struct {
	AgendaText string `json:"agenda_text"`
}

type analysisRealtimeMessage struct {
	Message model.RealtimeMessage `json:"message"`
}

type driftFeedbackPayload struct {
	SegmentID string `json:"segment_id"`
	IsDrift   bool   `json:"is_drift"`
}

func writeJSON(ctx context.Context, conn *websocket.Conn, v any) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, b)
}
