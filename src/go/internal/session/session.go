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
	Bigrams           *bigram.Database
	SequenceStartPos  int                  // Buffer position where current correct sequence started
	WordStartPos      int                  // Buffer position where current word starts
	HasSequenceErrors bool                 // True if current sequence has any wrong keystrokes
	AttemptStart      int                  // Index i Keystrokes där nuvarande ordförsök började
	CompletedAt       time.Time            // När sista tecknet skrevs
	Summary           stats.SessionSummary // Fryst resultat för sessionen
}

// WPM calculates words per minute (korrekta tecken)
func (s *Session) WPM() float64 {
	return s.wordsPerMinute(float64(s.CorrectChars()), time.Since(s.StartedAt).Seconds())
}

// RawWPM calculates raw words per minute (alla tangenttryck)
func (s *Session) RawWPM() float64 {
	return s.wordsPerMinute(float64(len(s.Keystrokes)), time.Since(s.StartedAt).Seconds())
}

// CorrectChars returns the number of characters typed correctly in the text.
// Efter en omstart räknas omstartade tecken inte igen (netto).
func (s *Session) CorrectChars() int {
	return s.Buffer.Cursor
}

// wordsPerMinute converts a number of characters to words per minute
func (s *Session) wordsPerMinute(chars, seconds float64) float64 {
	if seconds <= 0 {
		return 0
	}
	return chars / 5.0 * 60.0 / seconds
}

// WordsCompleted returns how many words that have been typed
func (s *Session) WordsCompleted() int {
	n := s.Buffer.Cursor
	if n <= 0 {
		return 0
	}
	if n > len(s.Buffer.Runes) {
		n = len(s.Buffer.Runes)
	}

	words := 1 // sista ordet räknas även om det inte följs av mellanslag
	for i := 0; i < n; i++ {
		if s.Buffer.Runes[i] == ' ' {
			words++
		}
	}
	return words
}

// Freeze beräknar sessionens resultat och fryser det, så att värdet inte
// ändras medan användaren tittar på sammanfattningen.
func (s *Session) Freeze() {
	end := s.CompletedAt
	if end.IsZero() {
		end = time.Now()
	}
	duration := end.Sub(s.StartedAt)
	seconds := duration.Seconds()

	s.Summary = stats.SessionSummary{
		Date:           s.StartedAt,
		WordsCompleted: s.WordsCompleted(),
		Errors:         s.TotalErrors,
		WPM:            s.wordsPerMinute(float64(s.CorrectChars()), seconds),
		RawWPM:         s.wordsPerMinute(float64(len(s.Keystrokes)), seconds),
		DurationMs:     duration.Milliseconds(),
		CharsTyped:     s.CorrectChars(),
	}
}

// Result returns the session result, freezing it first if needed
func (s *Session) Result() stats.SessionSummary {
	if s.Summary.Date.IsZero() {
		s.Freeze()
	}
	return s.Summary
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
