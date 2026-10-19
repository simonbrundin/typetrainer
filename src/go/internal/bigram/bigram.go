package bigram

// Stats tracks statistics for a single bigram
type Stats struct {
	Count   uint32 // Times encountered
	TotalMs uint64 // Total time in ms
	Errors  uint32 // Number of errors
}

// AvgMs returns average time per keystroke
func (s *Stats) AvgMs() float64 {
	if s.Count == 0 {
		return 0
	}
	return float64(s.TotalMs) / float64(s.Count)
}

// ErrorRate returns error rate as 0.0 to 1.0
func (s *Stats) ErrorRate() float64 {
	if s.Count == 0 {
		return 0
	}
	return float64(s.Errors) / float64(s.Count)
}

// Config for bigram analysis
type Config struct {
	MinSamples     uint32  // Min samples before tracking
	WeakThreshold  float64 // Error rate to be "weak"
	TimeWeight     float64
	ErrorWeight    float64
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		MinSamples:    10,
		WeakThreshold: 0.15,
		TimeWeight:    0.3,
		ErrorWeight:    0.7,
	}
}

// Database holds all bigram statistics
type Database struct {
	Stats  map[string]*Stats
	Config Config
}

// NewDatabase creates a new bigram database
func NewDatabase() *Database {
	return &Database{
		Stats:  make(map[string]*Stats),
		Config: DefaultConfig(),
	}
}

// Record records a keystroke for a bigram
func (db *Database) Record(bigram string, timeMs uint64, error bool) {
	if _, ok := db.Stats[bigram]; !ok {
		db.Stats[bigram] = &Stats{}
	}

	s := db.Stats[bigram]
	s.Count++
	s.TotalMs += timeMs
	if error {
		s.Errors++
	}
}

// GlobalAvgTime calculates global average time
func (db *Database) GlobalAvgTime() float64 {
	var totalCount, totalMs uint64
	for _, s := range db.Stats {
		totalCount += uint64(s.Count)
		totalMs += s.TotalMs
	}

	if totalCount == 0 {
		return 150.0 // Default assumed avg
	}
	return float64(totalMs) / float64(totalCount)
}

// DifficultyScore calculates difficulty (higher = needs practice)
func (s *Stats) DifficultyScore(avgTime float64, cfg *Config) float64 {
	if s.Count < cfg.MinSamples {
		return 0
	}

	errorComponent := s.ErrorRate() * cfg.ErrorWeight
	timeComponent := 0.0
	if avgTime > 0 {
		ratio := s.AvgMs() / avgTime
		if ratio > 3 {
			ratio = 3
		}
		timeComponent = ratio / 3 * cfg.TimeWeight
	}

	return errorComponent + timeComponent
}

// IsWeak returns true if bigram needs practice
func (s *Stats) IsWeak(avgTime float64, cfg *Config) bool {
	if s.Count < cfg.MinSamples {
		return false
	}

	return s.ErrorRate() >= cfg.WeakThreshold ||
		(avgTime > 0 && s.AvgMs() > avgTime*1.5)
}

// MergeBigrams merges another bigram database into this one
func (db *Database) MergeBigrams(other *Database) {
	for bigram, otherStats := range other.Stats {
		if existing, ok := db.Stats[bigram]; ok {
			existing.Count += otherStats.Count
			existing.TotalMs += otherStats.TotalMs
			existing.Errors += otherStats.Errors
		} else {
			db.Stats[bigram] = &Stats{
				Count:   otherStats.Count,
				TotalMs: otherStats.TotalMs,
				Errors:  otherStats.Errors,
			}
		}
	}
}

// BigramsInWord extracts all bigrams from a word
func BigramsInWord(word string) []string {
	runes := []rune(word)
	if len(runes) < 2 {
		return nil
	}

	result := make([]string, len(runes)-1)
	for i := 0; i < len(runes)-1; i++ {
		result[i] = string(runes[i]) + string(runes[i+1])
	}
	return result
}
