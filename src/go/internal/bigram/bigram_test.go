package bigram

import (
	"testing"
)

func TestNewDatabase(t *testing.T) {
	db := NewDatabase()
	if db == nil {
		t.Fatal("NewDatabase() returned nil")
	}
	if db.Stats == nil {
		t.Error("db.Stats should not be nil")
	}
	if db.Config.MinSamples != 10 {
		t.Errorf("Default MinSamples = %d, want 10", db.Config.MinSamples)
	}
}

func TestStats_AvgMs(t *testing.T) {
	tests := []struct {
		name    string
		stats   Stats
		wantAvg float64
	}{
		{
			name:    "zero count",
			stats:   Stats{Count: 0, TotalMs: 0},
			wantAvg: 0,
		},
		{
			name:    "single sample",
			stats:   Stats{Count: 1, TotalMs: 100},
			wantAvg: 100,
		},
		{
			name:    "multiple samples",
			stats:   Stats{Count: 10, TotalMs: 1000},
			wantAvg: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.AvgMs()
			if got != tt.wantAvg {
				t.Errorf("AvgMs() = %v, want %v", got, tt.wantAvg)
			}
		})
	}
}

func TestStats_ErrorRate(t *testing.T) {
	tests := []struct {
		name     string
		stats    Stats
		wantRate float64
	}{
		{
			name:     "zero count",
			stats:    Stats{Count: 0, Errors: 0},
			wantRate: 0,
		},
		{
			name:     "no errors",
			stats:    Stats{Count: 10, Errors: 0},
			wantRate: 0,
		},
		{
			name:     "50% error rate",
			stats:    Stats{Count: 10, Errors: 5},
			wantRate: 0.5,
		},
		{
			name:     "100% error rate",
			stats:    Stats{Count: 10, Errors: 10},
			wantRate: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.ErrorRate()
			if got != tt.wantRate {
				t.Errorf("ErrorRate() = %v, want %v", got, tt.wantRate)
			}
		})
	}
}

func TestDatabase_Record(t *testing.T) {
	db := NewDatabase()

	// Record a bigram
	db.Record("th", 100, false)

	if db.Stats["th"] == nil {
		t.Fatal("Bigram 'th' should be recorded")
	}

	if db.Stats["th"].Count != 1 {
		t.Errorf("Count = %d, want 1", db.Stats["th"].Count)
	}
	if db.Stats["th"].TotalMs != 100 {
		t.Errorf("TotalMs = %d, want 100", db.Stats["th"].TotalMs)
	}
	if db.Stats["th"].Errors != 0 {
		t.Errorf("Errors = %d, want 0", db.Stats["th"].Errors)
	}

	// Record with error
	db.Record("th", 200, true)

	if db.Stats["th"].Count != 2 {
		t.Errorf("Count = %d, want 2", db.Stats["th"].Count)
	}
	if db.Stats["th"].Errors != 1 {
		t.Errorf("Errors = %d, want 1", db.Stats["th"].Errors)
	}
}

func TestDatabase_GlobalAvgTime(t *testing.T) {
	db := NewDatabase()

	// Empty database
	avg := db.GlobalAvgTime()
	if avg != 150.0 {
		t.Errorf("GlobalAvgTime() = %v, want 150.0 (default)", avg)
	}

	// Add some bigrams
	db.Record("ab", 100, false)
	db.Record("ab", 200, false)
	db.Record("cd", 150, false)

	// (100 + 200 + 150) / 3 = 150
	avg = db.GlobalAvgTime()
	if avg != 150.0 {
		t.Errorf("GlobalAvgTime() = %v, want 150.0", avg)
	}
}

func TestStats_DifficultyScore(t *testing.T) {
	cfg := DefaultConfig()
	avgTime := 100.0

	tests := []struct {
		name    string
		stats   Stats
		minWant float64 // minimum expected value
	}{
		{
			name:    "insufficient samples",
			stats:   Stats{Count: 5, TotalMs: 200, Errors: 1},
			minWant: 0, // Should be 0 due to insufficient samples
		},
		{
			name:    "high error rate",
			stats:   Stats{Count: 20, TotalMs: 200, Errors: 10},
			minWant: 0.35, // At least 0.35 (50% error rate * 0.7 weight)
		},
		{
			name:    "normal performance",
			stats:   Stats{Count: 20, TotalMs: 100, Errors: 0},
			minWant: 0, // No errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.DifficultyScore(avgTime, &cfg)
			if got < tt.minWant {
				t.Errorf("DifficultyScore() = %v, want >= %v", got, tt.minWant)
			}
		})
	}
}

func TestStats_IsWeak(t *testing.T) {
	cfg := DefaultConfig()
	avgTime := 100.0

	tests := []struct {
		name   string
		stats  Stats
		isWeak bool
	}{
		{
			name:   "insufficient samples",
			stats:  Stats{Count: 5, TotalMs: 300, Errors: 3},
			isWeak: false, // Not enough samples
		},
		{
			name:   "high error rate",
			stats:  Stats{Count: 20, TotalMs: 100, Errors: 5},
			isWeak: true, // 25% error rate > 15% threshold
		},
		{
			name:   "very slow but accurate",
			stats:  Stats{Count: 20, TotalMs: 160 * 20, Errors: 0},
			isWeak: true, // 160ms avg > 150ms (1.5 * 100)
		},
		{
			name:   "normal performance",
			stats:  Stats{Count: 20, TotalMs: 100, Errors: 0},
			isWeak: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.IsWeak(avgTime, &cfg)
			if got != tt.isWeak {
				t.Errorf("IsWeak() = %v, want %v", got, tt.isWeak)
			}
		})
	}
}

func TestBigramsInWord(t *testing.T) {
	tests := []struct {
		word string
		want []string
	}{
		{"hej", []string{"he", "ej"}},
		{"ab", []string{"ab"}},
		{"a", nil},
		{"", nil},
		{"svensk", []string{"sv", "ve", "en", "ns", "sk"}},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := BigramsInWord(tt.word)
			if len(got) != len(tt.want) {
				t.Errorf("BigramsInWord(%q) len = %d, want %d", tt.word, len(got), len(tt.want))
				return
			}
			for i, w := range tt.want {
				if got[i] != w {
					t.Errorf("BigramsInWord(%q)[%d] = %q, want %q", tt.word, i, got[i], w)
				}
			}
		})
	}
}

func TestDatabase_MergeBigrams(t *testing.T) {
	db1 := NewDatabase()
	db1.Record("ab", 100, false)
	db1.Record("ab", 150, false)
	db1.Record("cd", 200, true)

	db2 := NewDatabase()
	db2.Record("ab", 100, false)
	db2.Record("ef", 300, false)

	db1.MergeBigrams(db2)

	if db1.Stats["ab"].Count != 3 {
		t.Errorf("Merged 'ab' Count = %d, want 3", db1.Stats["ab"].Count)
	}
	if db1.Stats["ab"].TotalMs != 350 {
		t.Errorf("Merged 'ab' TotalMs = %d, want 350", db1.Stats["ab"].TotalMs)
	}
	if db1.Stats["cd"].Count != 1 {
		t.Errorf("'cd' Count = %d, want 1", db1.Stats["cd"].Count)
	}
	if db1.Stats["ef"].Count != 1 {
		t.Errorf("'ef' Count = %d, want 1", db1.Stats["ef"].Count)
	}
}
