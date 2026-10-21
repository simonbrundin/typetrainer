package session

import (
	"fmt"

	"typetrainer/internal/stats"
	"typetrainer/internal/wordlist"

	"github.com/charmbracelet/bubbletea"
)

// restartKeyType är en speciell nyckeltyp för att starta om ordet (via ctrl+backspace)
const restartKeyType tea.KeyType = 0x1000

// RunSession starts a typing session
func RunSession(
	forcedWords []*wordlist.Word,
	wl *wordlist.WordList,
	s *stats.Stats,
	cfg stats.Config,
	mode Mode,
) {
	words := prepareSessionWords(mode, forcedWords, wl, s, cfg)
	if words == nil {
		return
	}

	session := newSession(words, mode, &cfg)
	model := createModel(session, wl, words, s, cfg, mode)

	runProgram(model)
	saveSessionResults(model, session, s)
}

func prepareSessionWords(
	mode Mode,
	forcedWords []*wordlist.Word,
	wl *wordlist.WordList,
	s *stats.Stats,
	cfg stats.Config,
) []*wordlist.Word {
	var words []*wordlist.Word

	if mode == ModePracticeBigram {
		if len(forcedWords) == 0 {
			fmt.Println("Inga ord hittades med detta bigram.")
			return nil
		}
		words = forcedWords
	} else {
		words = generateWords(mode, wl, s, cfg)
	}

	if len(words) == 0 {
		fmt.Println("Inga ord tillgängliga.")
		return nil
	}

	return words
}

func createModel(
	session *Session,
	wl *wordlist.WordList,
	words []*wordlist.Word,
	s *stats.Stats,
	cfg stats.Config,
	mode Mode,
) *Model {
	return &Model{
		session:   session,
		config:    &cfg,
		mode:      mode,
		wordlist:  wl,
		words:     words,
		showIntro: false, // Start directly without intro
		introDone: true,  // Mark intro as done for stats
		timeLeft:  cfg.SprintDuration,
		allStats:  s,
	}
}

func runProgram(model *Model) {
	p := tea.NewProgram(model,
		tea.WithFilter(func(_ tea.Model, msg tea.Msg) tea.Msg {
			// Konvertera ctrl+h och ctrl+w (från ctrl+backspace via kanata) till restartKeyType
			if key, ok := msg.(tea.KeyMsg); ok {
				if key.Type == tea.KeyCtrlH || key.Type == tea.KeyCtrlW {
					return tea.KeyMsg{Type: restartKeyType}
				}
			}
			return msg
		}),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}

func saveSessionResults(model *Model, session *Session, s *stats.Stats) {
	if !shouldSaveResults(model) {
		return
	}

	summary := createSessionSummary(session)
	mergeBigramStats(session, s)
	s.AddSession(summary)
	s.Save()
	s.PrintStats()
}

func shouldSaveResults(model *Model) bool {
	return model.introDone || model.done || model.showContinue
}

func createSessionSummary(session *Session) stats.SessionSummary {
	return session.Result()
}

func mergeBigramStats(session *Session, s *stats.Stats) {
	s.Bigrams.MergeBigrams(session.Bigrams)
}
