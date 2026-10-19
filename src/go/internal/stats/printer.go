package stats

import (
	"fmt"
)

// PrintStats prints statistics to stdout
func (s *Stats) PrintStats() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║          TYPE TRAINER STATS              ║")
	fmt.Println("╠══════════════════════════════════════════╣")

	// Total practice
	var totalWords, totalErrors int
	var totalTime int64
	for _, sess := range s.Sessions {
		totalWords += sess.WordsCompleted
		totalErrors += sess.Errors
		totalTime += sess.DurationMs
	}

	fmt.Printf("║  Totalt tränade ord:     %12d     ║\n", totalWords)
	fmt.Printf("║  Totalt antal fel:       %12d     ║\n", totalErrors)
	fmt.Printf("║  Total träningstid:      %12ds     ║\n", totalTime/1000)
	fmt.Println("║                                          ║")

	// Average WPM
	avgWPM := s.AvgWPM(10)
	avgRawWPM := s.AvgRawWPM(10)
	fmt.Printf("║  Snitt Raw WPM (10):     %12.1f     ║\n", avgRawWPM)
	fmt.Printf("║  Snitt WPM (10):         %12.1f     ║\n", avgWPM)

	// Personal best
	if pb, ok := s.PersonalBest["default"]; ok && pb > 0 {
		fmt.Printf("║  Personligt rekord:      %12.1f     ║\n", pb)
	}

	fmt.Println("║                                          ║")
	fmt.Printf("║  Antal sessioner:       %12d     ║\n", len(s.Sessions))

	fmt.Println("╠══════════════════════════════════════════╣")

	// Worst bigrams
	s.printWorstBigrams()

	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
}

func (s *Stats) printWorstBigrams() {
	avgTime := s.Bigrams.GlobalAvgTime()

	type bigramInfo struct {
		name       string
		errorRate  float64
		avgMs      float64
		difficulty float64
	}

	var worst []bigramInfo
	for name, stats := range s.Bigrams.Stats {
		if stats.Count >= s.Bigrams.Config.MinSamples {
			diff := stats.DifficultyScore(avgTime, &s.Bigrams.Config)
			worst = append(worst, bigramInfo{
				name:       name,
				errorRate:  stats.ErrorRate() * 100,
				avgMs:      stats.AvgMs(),
				difficulty: diff,
			})
		}
	}

	// Sort by difficulty (bubble sort - simple for small lists)
	for i := 0; i < len(worst)-1; i++ {
		for j := i + 1; j < len(worst); j++ {
			if worst[j].difficulty > worst[i].difficulty {
				worst[i], worst[j] = worst[j], worst[i]
			}
		}
	}

	if len(worst) == 0 {
		fmt.Println("║  Träna mer för att se svaga bigrams!     ║")
		return
	}

	fmt.Println("║  Svåraste bigrammen:                      ║")
	count := 5
	if count > len(worst) {
		count = len(worst)
	}
	for i := 0; i < count; i++ {
		w := worst[i]
		fmt.Printf("║    %d. %-6s  fel: %5.1f%%  tid: %5.0fms      ║\n",
			i+1, w.name, w.errorRate, w.avgMs)
	}
}
