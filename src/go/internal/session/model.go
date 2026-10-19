package session

import (
	"time"

	"typetrainer/internal/stats"
	"typetrainer/internal/wordlist"

	"github.com/charmbracelet/bubbletea"
)

// Model is the bubbletea model for the typing trainer
type Model struct {
	session      *Session
	config       *stats.Config
	mode         Mode
	wordlist     *wordlist.WordList
	words        []*wordlist.Word
	timeLeft     int
	timer        *time.Timer
	done         bool
	showIntro    bool
	introDone    bool
	showContinue bool // After session complete, ask if continue
	allStats     *stats.Stats
	escapeSeq    string // Buffer for escape sequence detection
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		// Handle resize (no-op, bubbletea handles it)

	case *time.Timer:
		m.timeLeft--
		if m.timeLeft <= 0 {
			m.showContinue = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// handleKeyPress handles all keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Intro screen - any key starts
	if m.showIntro {
		m.showIntro = false
		m.introDone = true
		return m, nil
	}

	// After session complete - Enter to continue, Esc to quit
	if m.showContinue {
		return m.handleContinuePrompt(msg)
	}

	// Quit on Ctrl+C
	if msg.Type == tea.KeyCtrlC {
		m.done = true
		return m, tea.Quit
	}

	// Session done
	if m.done {
		return m, tea.Quit
	}

	// Handle restart keys (Tab, Ctrl+U, Backspace, Ctrl+H)
	if m.isRestartKey(msg) {
		if m.session.HasSequenceErrors {
			m.restartSequence()
		}
		return m, nil
	}

	// If sequence has errors, only restart keys work
	if m.session.HasSequenceErrors {
		return m, nil
	}

	// Only handle printable characters
	return m.handleTypingInput(msg)
}

// isRestartKey checks if the key should restart the sequence
func (m *Model) isRestartKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyTab, tea.KeyCtrlU, tea.KeyBackspace, tea.KeyCtrlH, tea.KeyCtrlW, restartKeyType:
		return true
	}
	return false
}

// handleContinuePrompt handles input after session completes
func (m *Model) handleContinuePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		m.showContinue = false
		m.words = generateWords(m.mode, m.wordlist, m.allStats, *m.config)
		m.session = newSession(m.words, m.mode, m.config)
		return m, nil
	}
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}
	return m, nil
}

// restartSequence resets the cursor to the start of the current word
func (m *Model) restartSequence() {
	m.session.Buffer.Cursor = m.session.WordStartPos
	m.session.Keystrokes = nil
	m.session.HasSequenceErrors = false
	m.session.WordStartTime = time.Now()
}

// handleTypingInput processes a regular character input
func (m *Model) handleTypingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	runes := []rune(msg.String())
	if len(runes) != 1 {
		return m, nil
	}
	c := runes[0]

	buf := m.session.Buffer

	// Check if complete
	if buf.IsComplete() {
		m.showContinue = true
		if m.timer != nil {
			m.timer.Stop()
		}
		return m, nil
	}

	expected := buf.CurrentRune()
	timeMs := uint64(time.Since(m.session.WordStartTime).Milliseconds())

	// Get previous char for bigram tracking
	prev := m.getPreviousRune()

	correct := c == expected

	// Record keystroke
	m.session.Keystrokes = append(m.session.Keystrokes, Keystroke{
		Expected:   expected,
		Actual:     c,
		TimeMs:     timeMs,
		Correct:    correct,
		BigramPrev: prev,
	})

	// Count errors
	if !correct {
		m.session.TotalErrors++
		m.session.HasSequenceErrors = true
	}

	// Only advance cursor on correct keypress
	if correct {
		m.recordBigramProgress(prev, expected, timeMs)
		buf.Cursor++
		m.session.CorrectChars++ // Track total correct chars (doesn't decrease on restart)
		m.session.WordStartTime = time.Now()

		// Update WordStartPos when entering a new word (after space)
		if buf.Cursor < buf.Len() && buf.Runes[buf.Cursor] == ' ' {
			// Skip the space and set WordStartPos to the next word
			m.session.WordStartPos = buf.Cursor + 1
		}

		if buf.IsComplete() {
			m.showContinue = true
			if m.timer != nil {
				m.timer.Stop()
			}
		}
	}

	return m, nil
}

// getPreviousRune returns the previous expected rune for bigram tracking
func (m *Model) getPreviousRune() rune {
	if len(m.session.Keystrokes) == 0 {
		return 0
	}
	return m.session.Keystrokes[len(m.session.Keystrokes)-1].Expected
}

// recordBigramProgress records bigram statistics for correct keystroke
func (m *Model) recordBigramProgress(prev, expected rune, timeMs uint64) {
	if prev == 0 {
		return
	}
	bigramStr := string(prev) + string(expected)
	m.session.Bigrams.Record(bigramStr, timeMs, false)
}

// View implements tea.Model
func (m *Model) View() string {
	if m.showIntro {
		return m.introView()
	}
	if m.showContinue || m.done {
		return m.summaryView()
	}
	return m.sessionView()
}
