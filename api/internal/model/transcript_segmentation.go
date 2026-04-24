package model

import (
	"fmt"
	"strings"
)

// Segment is an analysis chunk derived from a contiguous range of transcript turns.
//
// Indices are 0-based and inclusive.
//
// segment_id is stable as long as the transcript ordering and segmentation options are unchanged.
// For MVP, it is derived from the source turn index range.
type Segment struct {
	SegmentID           string             `json:"segment_id"`
	StartTurnIdx        int                `json:"start_turn_idx"`
	EndTurnIdx          int                `json:"end_turn_idx"`
	Text                string             `json:"text"`
	SpeakerDistribution map[string]float64 `json:"speaker_distribution,omitempty"`
}

// SegmentOptions configures how turns are grouped into segments.
//
// MaxTokens and MaxChars are soft caps; if a single turn exceeds a cap, it will still
// be placed into its own segment.
type SegmentOptions struct {
	// MaxTokens is an approximate token cap based on whitespace-separated fields.
	// 0 means "no token cap".
	MaxTokens int
	// MaxChars is a character/byte cap based on len(trimmed turn text).
	// 0 means "no char cap".
	MaxChars int

	// IncludeSpeakerLabels prefixes each turn line with "Speaker: ".
	IncludeSpeakerLabels bool

	// ComputeSpeakerDistribution includes a per-segment speaker distribution.
	// The distribution is based on character counts of each speaker's turn texts.
	ComputeSpeakerDistribution bool
}

func (o SegmentOptions) normalized() SegmentOptions {
	// Clamp invalid caps; booleans are used as provided.
	if o.MaxTokens < 0 {
		o.MaxTokens = 0
	}
	if o.MaxChars < 0 {
		o.MaxChars = 0
	}
	return o
}

// SegmentTranscript groups transcript turns into deterministic analysis segments.
func SegmentTranscript(t Transcript, opts SegmentOptions) []Segment {
	opts = opts.normalized()

	if len(t.Turns) == 0 {
		return nil
	}

	var out []Segment

	curStart := -1
	curEnd := -1
	curTokens := 0
	curChars := 0
	var curLines []string
	var speakerChars map[string]int

	flush := func() {
		if curStart < 0 || curEnd < 0 {
			return
		}
		text := strings.TrimSpace(strings.Join(curLines, "\n"))
		if text == "" {
			// Never output empty segments.
			curStart, curEnd, curTokens, curChars, curLines, speakerChars = -1, -1, 0, 0, nil, nil
			return
		}

		seg := Segment{
			SegmentID:    fmt.Sprintf("seg-%d-%d", curStart, curEnd),
			StartTurnIdx: curStart,
			EndTurnIdx:   curEnd,
			Text:         text,
		}

		if opts.ComputeSpeakerDistribution {
			total := 0
			for _, c := range speakerChars {
				total += c
			}
			if total > 0 {
				dist := make(map[string]float64, len(speakerChars))
				for spk, c := range speakerChars {
					dist[spk] = float64(c) / float64(total)
				}
				seg.SpeakerDistribution = dist
			}
		}

		out = append(out, seg)
		curStart, curEnd, curTokens, curChars, curLines, speakerChars = -1, -1, 0, 0, nil, nil
	}

	for i, turn := range t.Turns {
		trimmed := strings.TrimSpace(turn.Text)
		turnTokens := countApproxTokens(trimmed)
		turnChars := len(trimmed)

		wouldExceed := false
		if curStart >= 0 {
			if opts.MaxTokens > 0 && curTokens+turnTokens > opts.MaxTokens {
				wouldExceed = true
			}
			if opts.MaxChars > 0 && curChars+turnChars > opts.MaxChars {
				wouldExceed = true
			}
		}

		if wouldExceed {
			flush()
		}

		if curStart < 0 {
			curStart = i
			curEnd = i
			curTokens = 0
			curChars = 0
			curLines = nil
			if opts.ComputeSpeakerDistribution {
				speakerChars = make(map[string]int)
			}
		}

		curEnd = i
		curTokens += turnTokens
		curChars += turnChars

		if trimmed != "" {
			curLines = append(curLines, renderTurnLine(turn, opts.IncludeSpeakerLabels))
			if opts.ComputeSpeakerDistribution {
				spk := "<unknown>"
				if turn.Speaker != nil && strings.TrimSpace(*turn.Speaker) != "" {
					spk = strings.TrimSpace(*turn.Speaker)
				}
				speakerChars[spk] += turnChars
			}
		}
	}

	flush()
	return out
}

func countApproxTokens(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Fields(s))
}

func renderTurnLine(t TranscriptTurn, includeSpeaker bool) string {
	text := strings.TrimSpace(t.Text)
	if !includeSpeaker {
		return text
	}
	if t.Speaker == nil || strings.TrimSpace(*t.Speaker) == "" {
		return text
	}
	return strings.TrimSpace(*t.Speaker) + ": " + text
}

// RenderSegmentsPlaintext renders segments with visible boundaries.
// This is intended for debugging and local inspection (not a stable API surface).
func RenderSegmentsPlaintext(segs []Segment) string {
	if len(segs) == 0 {
		return ""
	}
	var b strings.Builder
	for i, s := range segs {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("--- ")
		b.WriteString(s.SegmentID)
		b.WriteString(" turns[")
		b.WriteString(fmt.Sprintf("%d..%d", s.StartTurnIdx, s.EndTurnIdx))
		b.WriteString("] ---\n")
		b.WriteString(s.Text)
	}
	return b.String()
}
