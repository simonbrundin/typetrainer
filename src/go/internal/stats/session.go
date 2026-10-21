package stats

import (
	"sort"
	"time"

	"typetrainer/internal/bigram"
	"typetrainer/internal/wordlist"
)

// MaxSessions is the maximum number of sessions to keep
const MaxSessions = 100

// SessionSummary holds summary of a completed session
type SessionSummary struct {
	Date           time.Time `json:"date"`
	WordsCompleted int       `json:"words_completed"`
	Errors         int       `json:"errors"`
	WPM            float64   `json:"wpm"`
	RawWPM         float64   `json:"raw_wpm"`
	DurationMs     int64     `json:"duration_ms"`
	CharsTyped     int       `json:"chars_typed"`
}

// Accuracy calculates accuracy percentage
func (s *SessionSummary) Accuracy() float64 {
	chars := s.CharsTyped
	if chars == 0 {
		// Äldre sessioner har inte teckenantal, approximera då med 5 tecken/ord
		chars = s.WordsCompleted * 5
	}
	if chars == 0 {
		return 100
	}
	if s.Errors > chars {
		return 0
	}
	return float64(chars-s.Errors) / float64(chars) * 100
}

// Score returns WPM * accuracy / 100 for a session
func (s *SessionSummary) Score() float64 {
	return s.WPM * s.Accuracy() / 100
}

// Stats holds all persistent statistics
type Stats struct {
	Bigrams      *bigram.Database   `json:"bigrams"`
	Sessions     []SessionSummary   `json:"sessions"`
	CustomWords  []string           `json:"custom_words"`
	PersonalBest map[string]float64 `json:"personal_best"`
}

// New creates new stats
func New() *Stats {
	return &Stats{
		Bigrams:      bigram.NewDatabase(),
		Sessions:     make([]SessionSummary, 0),
		CustomWords:  make([]string, 0),
		PersonalBest: make(map[string]float64),
	}
}

// AddSession adds a completed session
func (s *Stats) AddSession(summary SessionSummary) {
	// Update personal best
	if wpm := summary.WPM; wpm > s.PersonalBest["default"] {
		s.PersonalBest["default"] = wpm
	}

	s.Sessions = append(s.Sessions, summary)

	// Keep last MaxSessions sessions
	if len(s.Sessions) > MaxSessions {
		s.Sessions = s.Sessions[len(s.Sessions)-MaxSessions:]
	}
}

// RecentSessions returns last n sessions
func (s *Stats) RecentSessions(n int) []SessionSummary {
	if len(s.Sessions) <= n {
		return s.Sessions
	}
	return s.Sessions[len(s.Sessions)-n:]
}

// AvgWPM calculates average WPM over last n sessions
func (s *Stats) AvgWPM(n int) float64 {
	recent := s.RecentSessions(n)
	if len(recent) == 0 {
		return 0
	}

	var sum float64
	for _, sess := range recent {
		sum += sess.WPM
	}
	return sum / float64(len(recent))
}

// AvgRawWPM calculates average Raw WPM over last n sessions
func (s *Stats) AvgRawWPM(n int) float64 {
	recent := s.RecentSessions(n)
	if len(recent) == 0 {
		return 0
	}

	var sum float64
	for _, sess := range recent {
		sum += sess.RawWPM
	}
	return sum / float64(len(recent))
}

// SessionScores returns scores for all sessions
func (s *Stats) SessionScores() []float64 {
	scores := make([]float64, len(s.Sessions))
	for i, sess := range s.Sessions {
		scores[i] = sess.Score()
	}
	return scores
}

// AvgScore calculates average score (WPM * accuracy) over last n sessions
func (s *Stats) AvgScore(n int) float64 {
	recent := s.RecentSessions(n)
	if len(recent) == 0 {
		return 0
	}

	var sum float64
	for _, sess := range recent {
		sum += sess.Score()
	}
	return sum / float64(len(recent))
}

// Reset clears all statistics
func (s *Stats) Reset() {
	s.Bigrams = bigram.NewDatabase()
	s.Sessions = make([]SessionSummary, 0)
	s.PersonalBest = make(map[string]float64)
}

// WeakBigrams returns bigrams that need practice
func (s *Stats) WeakBigrams(wl *wordlist.WordList, cfg *bigram.Config) []string {
	avgTime := s.Bigrams.GlobalAvgTime()

	type bigramScore struct {
		bigram string
		score  float64
	}

	var weak []bigramScore
	for bgram, stats := range s.Bigrams.Stats {
		if stats.IsWeak(avgTime, cfg) {
			score := stats.DifficultyScore(avgTime, cfg)
			weak = append(weak, bigramScore{bgram, score})
		}
	}

	if len(weak) == 0 {
		return nil
	}

	// Sort by score descending
	sort.Slice(weak, func(i, j int) bool {
		return weak[i].score > weak[j].score
	})

	result := make([]string, len(weak))
	for i, w := range weak {
		result[i] = w.bigram
	}
	return result
}
