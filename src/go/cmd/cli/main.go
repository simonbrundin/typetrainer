package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"typetrainer/internal/session"
	"typetrainer/internal/stats"
	"typetrainer/internal/wordlist"
)

var (
	wordsFlag    int
	durationFlag int
	bigramFlag   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "typetrainer",
		Short: "Swedish typing trainer with adaptive bigram practice",
		Long:  `A CLI typing trainer that adapts to your weaknesses by tracking and prioritizing
practice on difficult two-letter combinations (bigrams).`,
	}

	// Adaptive mode (default)
	adaptiveCmd := &cobra.Command{
		Use:   "adaptive",
		Short: "Adaptive training focusing on weak bigrams",
		Run:   runAdaptive,
	}
	adaptiveCmd.Flags().IntVarP(&wordsFlag, "words", "w", 20, "Number of words per session")

	// Random mode
	randomCmd := &cobra.Command{
		Use:   "random",
		Short: "Random word practice",
		Run:   runRandom,
	}
	randomCmd.Flags().IntVarP(&wordsFlag, "words", "w", 20, "Number of words")

	// Sprint mode
	sprintCmd := &cobra.Command{
		Use:   "sprint [seconds]",
		Short: "Timed sprint challenge",
		Args:  cobra.RangeArgs(0, 1),
		Run:   runSprint,
	}
	sprintCmd.Flags().IntVarP(&durationFlag, "duration", "d", 60, "Duration in seconds")

	// Practice specific bigram
	practiceCmd := &cobra.Command{
		Use:   "practice <bigram>",
		Short: "Practice specific bigram",
		Args:  cobra.ExactArgs(1),
		Run:   runPractice,
	}

	// Stats
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show your statistics",
		Run:   runStats,
	}

	// Reset
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset all statistics",
		Run:   runReset,
	}

	rootCmd.AddCommand(adaptiveCmd, randomCmd, sprintCmd, practiceCmd, statsCmd, resetCmd)

	// Default to adaptive if no subcommand
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "adaptive")
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runAdaptive(cmd *cobra.Command, args []string) {
	cfg := stats.LoadConfig()
	cfg.WordsPerSession = wordsFlag

	wl := wordlist.New()
	s := stats.Load()
	weakBigrams := s.WeakBigrams(wl, &cfg.Bigram)
	practiceWords := wl.AdaptiveSelection(weakBigrams, cfg.WordsPerSession)

	session.RunSession(practiceWords, wl, s, cfg, session.ModeAdaptive)
}

func runRandom(cmd *cobra.Command, args []string) {
	cfg := stats.LoadConfig()
	cfg.WordsPerSession = wordsFlag

	wl := wordlist.New()
	s := stats.Load()

	session.RunSession(nil, wl, s, cfg, session.ModeRandom)
}

func runSprint(cmd *cobra.Command, args []string) {
	cfg := stats.LoadConfig()

	duration := durationFlag
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &duration)
	}

	cfg.SprintDuration = duration

	wl := wordlist.New()
	s := stats.Load()

	session.RunSession(nil, wl, s, cfg, session.ModeSprint)
}

func runPractice(cmd *cobra.Command, args []string) {
	bigram := args[0]
	cfg := stats.LoadConfig()

	wl := wordlist.New()
	s := stats.Load()

	words := wl.WordsWithBigram(bigram)

	session.RunSession(words, wl, s, cfg, session.ModePracticeBigram)
}

func runStats(cmd *cobra.Command, args []string) {
	s := stats.Load()
	s.PrintStats()
}

func runReset(cmd *cobra.Command, args []string) {
	fmt.Println("⚠️  Vill du verkligen nollställa ALL statistik?")
	fmt.Print("   Skriv 'ja' för att bekräfta: ")

	var input string
	fmt.Scanln(&input)

	if input == "ja" {
		s := stats.Load()
		s.Reset()
		s.Save()
		fmt.Println("✅ Statistiken har nollställts.")
	} else {
		fmt.Println("Avbröt.")
	}
}
