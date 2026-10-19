package stats

import (
	"typetrainer/internal/bigram"
)

// Config holds application configuration
type Config struct {
	WordsPerSession int            `json:"words_per_session"`
	ShowNextWords   int            `json:"show_next_words"`
	SprintDuration  int           `json:"sprint_duration"`
	Bigram          bigram.Config  `json:"bigram"`
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		WordsPerSession: 20,
		ShowNextWords:   3,
		SprintDuration:  60,
		Bigram:         bigram.DefaultConfig(),
	}
}
