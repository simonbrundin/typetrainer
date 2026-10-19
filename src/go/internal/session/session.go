package session

import (
	"strings"
	"time"

	"typetrainer/internal/bigram"
	"typetrainer/internal/stats"
	"typetrainer/internal/wordlist"
)

// restartKeyCodes är de tangentkoder som startar om ordet
var restartKeyCodes = map[string]bool{
	"backspace": true,
	"ctrl+u":    true,
	"ctrl+h":    true,
	"tab":       true,
	// ctrl+backspace kan komma som escape-sekvens
	"esc[3;5~": true,
	"esc[3~":   true,
}

// TextBuffer holds the continuous text being typed
type TextBuffer struct {
	Text   string // All text to type
	Runes  []rune // Runes for proper Unicode handling
	Cursor int    // Current position (in runes)
}

// NewTextBuffer creates a new TextBuffer from words
func NewTextBuffer(words []*wordlist.Word) *TextBuffer {
	var parts []string
	for _, w := range words {
		parts = append(parts, w.Text)
	}
	text := strings.Join(parts, " ")
	return &TextBuffer{
		Text:   text,
		Runes:  []rune(text),
		Cursor: 0,
	}
}

// Len returns the total length in runes
func (tb *TextBuffer) Len() int {
	return len(tb.Runes)
}

// IsComplete returns true if all text is typed
func (tb *TextBuffer) IsComplete() bool {
	return tb.Cursor >= len(tb.Runes)
}

// Progress returns progress as 0.0 to 1.0
func (tb *TextBuffer) Progress() float64 {
	if len(tb.Runes) == 0 {
		return 1.0
	}
	return float64(tb.Cursor) / float64(len(tb.Runes))
}

// CurrentRune returns the current rune to type
func (tb *TextBuffer) CurrentRune() rune {
	if tb.Cursor >= len(tb.Runes) {
		return 0
	}
	return tb.Runes[tb.Cursor]
}

// Mode represents the practice mode
type Mode int

const (
	ModeAdaptive Mode = iota
	ModeRandom
	ModeSprint
	ModePracticeBigram
)

// String returns the display name for the mode
func (m Mode) String() string {
	switch m {
	case ModeAdaptive:
		return "Adaptiv träning"
	case ModeRandom:
		return "Slumpmässig träning"
	case ModeSprint:
		return "Sprint"
	case ModePracticeBigram:
		return "Bigram-träning"
	default:
		return "Okänt"
	}
}

// Keystroke represents a single keystroke
type Keystroke struct {
	Expected   rune
	Actual     rune
	TimeMs     uint64
	Correct    bool
	BigramPrev rune
}

// WordResult holds result for a word
type WordResult struct {
	Text       string
	Keystrokes []Keystroke
	Completed  bool
	Errors     int
}

// Accuracy calculates accuracy for this word
func (w *WordResult) Accuracy() float64 {
	if len(w.Keystrokes) == 0 {
		return 100
	}
	correct := 0
	for _, k := range w.Keystrokes {
		if k.Correct {
			correct++
		}
	}
	return float64(correct) / float64(len(w.Keystrokes)) * 100
}

// Session holds the current typing session state
type Session struct {
	Buffer            *TextBuffer
	Keystrokes        []Keystroke
	StartedAt         time.Time
	WordStartTime     time.Time
	TotalErrors       int
	CorrectChars      int  // Total correct characters typed (doesn't decrease on restart)
	Bigrams           *bigram.Database
	SequenceStartPos  int  // Buffer position where current correct sequence started
	WordStartPos      int  // Buffer position where current word starts
	HasSequenceErrors bool // True if current sequence has any wrong keystrokes
}

// WPM calculates words per minute (correct characters only)
func (s *Session) WPM() float64 {
	elapsed := time.Since(s.StartedAt).Seconds()
	if elapsed == 0 {
		return 0
	}
	// Use CorrectChars instead of Buffer.Cursor (which decreases on restart)
	return float64(s.CorrectChars) / 5.0 * 60.0 / elapsed
}

// RawWPM calculates raw words per minute (all keystrokes)
func (s *Session) RawWPM() float64 {
	elapsed := time.Since(s.StartedAt).Seconds()
	if elapsed == 0 {
		return 0
	}
	// Monkeytype: (all_chars / 5) * (60 / elapsed_seconds)
	return float64(len(s.Keystrokes)) / 5.0 * 60.0 / elapsed
}

// Accuracy calculates overall accuracy
func (s *Session) Accuracy() float64 {
	if len(s.Keystrokes) == 0 {
		return 100
	}
	errors := 0
	for _, k := range s.Keystrokes {
		if !k.Correct {
			errors++
		}
	}
	return float64(len(s.Keystrokes)-errors) / float64(len(s.Keystrokes)) * 100
}

// IsComplete checks if session is done
func (s *Session) IsComplete() bool {
	return s.Buffer.IsComplete()
}

// newSession creates a fresh session
func newSession(words []*wordlist.Word, mode Mode, config *stats.Config) *Session {
	buffer := NewTextBuffer(words)
	return &Session{
		Buffer:            buffer,
		Keystrokes:        make([]Keystroke, 0),
		StartedAt:         time.Now(),
		WordStartTime:     time.Now(),
		Bigrams:           bigram.NewDatabase(),
		SequenceStartPos:  0,
		WordStartPos:      0,
		HasSequenceErrors: false,
	}
}

// generateWords generates words based on mode and stats
func generateWords(mode Mode, wl *wordlist.WordList, s *stats.Stats, cfg stats.Config) []*wordlist.Word {
	switch mode {
	case ModeAdaptive:
		weak := s.WeakBigrams(wl, &cfg.Bigram)
		return wl.AdaptiveSelection(weak, cfg.WordsPerSession)
	case ModeRandom:
		return wl.RandomWords(cfg.WordsPerSession)
	case ModeSprint:
		return wl.RandomWords(cfg.SprintDuration * 3)
	default:
		return wl.RandomWords(cfg.WordsPerSession)
	}
}
