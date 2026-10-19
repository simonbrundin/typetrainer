package stats

import (
	"testing"
	"time"
)

func TestSessionSummary_Accuracy(t *testing.T) {
	tests := []struct {
		name     string
		summary  SessionSummary
		wantAcc  float64
	}{
		{
			name:    "zero words",
			summary: SessionSummary{WordsCompleted: 0, Errors: 0},
			wantAcc: 100,
		},
		{
			name:    "no errors",
			summary: SessionSummary{WordsCompleted: 10, Errors: 0},
			wantAcc: 100,
		},
		{
			name:    "with errors",
			summary: SessionSummary{WordsCompleted: 10, Errors: 2},
			wantAcc: 96, // (50-2)/50 * 100 = 96%
		},
		{
			name:    "all errors",
			summary: SessionSummary{WordsCompleted: 5, Errors: 25},
			wantAcc: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.summary.Accuracy()
			if got != tt.wantAcc {
				t.Errorf("Accuracy() = %v, want %v", got, tt.wantAcc)
			}
		})
	}
}

func TestSessionSummary_Score(t *testing.T) {
	summary := SessionSummary{
		WPM:            50,
		WordsCompleted: 10,
		Errors:         0,
	}

	// WPM * accuracy / 100 = 50 * 100 / 100 = 50
	want := 50.0
	got := summary.Score()
	if got != want {
		t.Errorf("Score() = %v, want %v", got, want)
	}
}

func TestNew(t *testing.T) {
	s := New()

	if s.Bigrams == nil {
		t.Error("Bigrams should not be nil")
	}
	if s.Sessions == nil {
		t.Error("Sessions should not be nil")
	}
	if len(s.Sessions) != 0 {
		t.Error("Sessions should be empty")
	}
	if s.PersonalBest == nil {
		t.Error("PersonalBest should not be nil")
	}
}

func TestStats_AddSession(t *testing.T) {
	s := New()

	session := SessionSummary{
		Date:           time.Now(),
		WordsCompleted: 10,
		WPM:            50,
	}

	s.AddSession(session)

	if len(s.Sessions) != 1 {
		t.Errorf("len(Sessions) = %d, want 1", len(s.Sessions))
	}
}

func TestStats_AddSession_UpdatesPersonalBest(t *testing.T) {
	s := New()

	s.AddSession(SessionSummary{WPM: 30})
	s.AddSession(SessionSummary{WPM: 50})
	s.AddSession(SessionSummary{WPM: 40})

	if s.PersonalBest["default"] != 50 {
		t.Errorf("PersonalBest[default] = %v, want 50", s.PersonalBest["default"])
	}
}

func TestStats_AddSession_LimitsSessions(t *testing.T) {
	s := New()

	// Add more than MaxSessions
	for i := 0; i < MaxSessions+50; i++ {
		s.AddSession(SessionSummary{WPM: float64(i)})
	}

	if len(s.Sessions) != MaxSessions {
		t.Errorf("len(Sessions) = %d, want %d", len(s.Sessions), MaxSessions)
	}
}

func TestStats_RecentSessions(t *testing.T) {
	s := New()

	for i := 0; i < 20; i++ {
		s.AddSession(SessionSummary{WPM: float64(i)})
	}

	tests := []struct {
		n    int
		want int
	}{
		{5, 5},
		{10, 10},
		{20, 20},
		{25, 20}, // More than available
	}

	for _, tt := range tests {
		recent := s.RecentSessions(tt.n)
		if len(recent) != tt.want {
			t.Errorf("RecentSessions(%d) len = %d, want %d", tt.n, len(recent), tt.want)
		}
	}
}

func TestStats_AvgWPM(t *testing.T) {
	s := New()

	// Add sessions with known WPM values
	for _, wpm := range []float64{50, 60, 70, 80, 90} {
		s.AddSession(SessionSummary{WPM: wpm})
	}

	tests := []struct {
		n    int
		want float64
	}{
		{1, 90},              // Last one
		{3, 80},              // (90+80+70)/3 = 80
		{10, 70},             // All 5, sum/5
	}

	for _, tt := range tests {
		got := s.AvgWPM(tt.n)
		if got != tt.want {
			t.Errorf("AvgWPM(%d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}

func TestStats_AvgWPM_Empty(t *testing.T) {
	s := New()

	if s.AvgWPM(10) != 0 {
		t.Error("AvgWPM on empty Stats should return 0")
	}
}

func TestStats_AvgRawWPM(t *testing.T) {
	s := New()

	for _, wpm := range []float64{40, 50, 60} {
		s.AddSession(SessionSummary{RawWPM: wpm})
	}

	got := s.AvgRawWPM(3)
	want := 50.0

	if got != want {
		t.Errorf("AvgRawWPM(3) = %v, want %v", got, want)
	}
}

func TestStats_SessionScores(t *testing.T) {
	s := New()

	s.AddSession(SessionSummary{
		WPM:            50,
		WordsCompleted: 10,
		Errors:         0,
	})

	scores := s.SessionScores()
	if len(scores) != 1 {
		t.Errorf("len(SessionScores()) = %d, want 1", len(scores))
	}
}

func TestStats_AvgScore(t *testing.T) {
	s := New()

	// Add sessions with known scores
	s.AddSession(SessionSummary{
		WPM:            50,
		WordsCompleted: 10,
		Errors:         0, // Score = 50
	})
	s.AddSession(SessionSummary{
		WPM:            40,
		WordsCompleted: 10,
		Errors:         0, // Score = 40
	})

	got := s.AvgScore(2)
	want := 45.0

	if got != want {
		t.Errorf("AvgScore(2) = %v, want %v", got, want)
	}
}

func TestStats_Reset(t *testing.T) {
	s := New()

	s.AddSession(SessionSummary{WPM: 50})
	s.AddSession(SessionSummary{WPM: 60})

	s.Reset()

	if len(s.Sessions) != 0 {
		t.Errorf("After Reset, len(Sessions) = %d, want 0", len(s.Sessions))
	}
	if len(s.PersonalBest) != 0 {
		t.Errorf("After Reset, len(PersonalBest) = %d, want 0", len(s.PersonalBest))
	}
}
