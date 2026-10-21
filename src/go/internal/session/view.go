package session

import (
	"fmt"
	"strings"

	"typetrainer/internal/stats"

	"github.com/charmbracelet/lipgloss"
)

// DisplayLength is how many characters to show at once
const DisplayLength = 50

// ProgressBarWidth is the width of the progress bar in characters
const ProgressBarWidth = 40

// Styles
var (
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	// Text styles
	currentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Reverse(true)
	typedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	remainingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	progressStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	borderStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	warningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

// boxWidth är den inre bredden på textrutorna
const boxWidth = 44

// renderBox ritar en ram runt rubrik och innehåll
func renderBox(title string, rows []string) string {
	var b strings.Builder

	b.WriteString(borderStyle.Render("+"+strings.Repeat("-", boxWidth)+"+") + "\n")
	b.WriteString(borderStyle.Render("|") + centerText(title, boxWidth) + borderStyle.Render("|") + "\n")
	b.WriteString(borderStyle.Render("+"+strings.Repeat("-", boxWidth)+"+") + "\n")

	for _, row := range rows {
		b.WriteString(borderStyle.Render("|") + padText(row, boxWidth) + borderStyle.Render("|") + "\n")
	}

	b.WriteString(borderStyle.Render("+"+strings.Repeat("-", boxWidth)+"+") + "\n")
	return b.String()
}

// centerText centrerar text i en given bredd
func centerText(text string, width int) string {
	runes := []rune(text)
	if len(runes) >= width {
		return string(runes[:width])
	}
	left := (width - len(runes)) / 2
	right := width - len(runes) - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

// padText vänsterjusterar text i en given bredd
func padText(text string, width int) string {
	runes := []rune(text)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return text + strings.Repeat(" ", width-len(runes))
}

// introView renders the intro screen
func (m *Model) introView() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(renderBox("TYPE TRAINER - "+m.mode.String(), []string{
		"  Skriv orden som visas.",
		"  Tryck Ctrl+C för att avsluta.",
	}))
	b.WriteString("\n  Tryck Enter for att börja...")
	return b.String()
}

// sessionView renders the active typing session
func (m *Model) sessionView() string {
	var b strings.Builder

	s := m.session
	buf := s.Buffer
	text := buf.Text
	cursor := buf.Cursor

	b.WriteString("\n\n\n") // Extra space above

	// Convert to runes for proper Unicode handling
	runes := []rune(text)
	totalLen := len(runes)

	// Display settings
	const leftPadding = 5
	const historyLen = 8
	const charSpacing = 0 // No extra spaces between characters

	// Calculate visible range
	startRune := cursor - historyLen
	if startRune < 0 {
		startRune = 0
	}

	// NEVER show leading spaces
	for startRune < cursor && startRune < totalLen {
		if runes[startRune] != ' ' {
			break
		}
		startRune++
	}
	if startRune >= cursor {
		startRune = cursor - 1
		if startRune < 0 {
			startRune = 0
		}
	}

	endRune := cursor + DisplayLength - historyLen
	if endRune > totalLen {
		endRune = totalLen
	}

	// Left padding
	b.WriteString(strings.Repeat(" ", leftPadding))

	// Show text with spacing between characters
	for i := startRune; i < endRune; i++ {
		r := runes[i]
		runeStr := string(r)

		if i < cursor {
			// Already typed - mörkgrå
			b.WriteString(typedStyle.Render(runeStr))
		} else if i == cursor {
			// Current position - gul, fet, reverse
			b.WriteString(currentStyle.Render(runeStr))
		} else {
			// Remaining - vit, fet
			b.WriteString(remainingStyle.Render(runeStr))
		}
		// Add spacing after each character
		b.WriteString(strings.Repeat(" ", charSpacing))
	}

	b.WriteString("\n\n")

	// Pad with spaces to fill display
	displayedLen := endRune - startRune + leftPadding
	for i := displayedLen; i < DisplayLength+leftPadding; i++ {
		b.WriteString(" ")
	}

	// Stats line
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  WPM: %5.1f  │  Accuracy: %5.1f%%  │  Errors: %d",
		s.WPM(), s.Accuracy(), s.TotalErrors))

	if m.mode == ModeSprint {
		b.WriteString(fmt.Sprintf("  │  Time: %ds", m.timeLeft))
	}

	// Progress bar
	b.WriteString("\n\n  [")
	progress := buf.Progress() * ProgressBarWidth
	for i := 0; i < ProgressBarWidth; i++ {
		if float64(i) < progress {
			b.WriteString(progressStyle.Render("█"))
		} else {
			b.WriteString(dimStyle.Render("░"))
		}
	}
	b.WriteString("]")
	b.WriteString(fmt.Sprintf(" %.0f%%", buf.Progress()*100))

	// Show hint if sequence has errors
	if s.HasSequenceErrors {
		b.WriteString(warningStyle.Render("  Tab/Ctrl+Backspace = starta om"))
	}

	return b.String()
}

// summaryView renders the session summary
func (m *Model) summaryView() string {
	s := m.session
	result := s.Result()
	wpm := result.WPM
	rawWpm := result.RawWPM
	accuracy := s.Accuracy()
	charsTyped := s.CorrectChars()

	msg := getPerformanceMessage(wpm, accuracy)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(renderBox("SESSION COMPLETE!", []string{
		fmt.Sprintf("  Tecken skrivna:  %10d", charsTyped),
		fmt.Sprintf("  Ord:             %10d", result.WordsCompleted),
		fmt.Sprintf("  Totalt fel:      %10d", s.TotalErrors),
		fmt.Sprintf("  Raw WPM:         %10.1f", rawWpm),
		fmt.Sprintf("  WPM (korrekt):   %10.1f", wpm),
		fmt.Sprintf("  Accuracy:        %10.1f%%", accuracy),
		"",
		fmt.Sprintf("  %s", msg),
	}))

	// Add progress graph if we have stats
	if m.allStats != nil {
		graph := m.renderSessionGraph(wpm)
		if graph != "" {
			b.WriteString("\n")
			b.WriteString(graph)
		}
	}

	if m.showContinue {
		b.WriteString(dimStyle.Render("  Enter = kor igen   |   Esc = avsluta"))
	} else {
		b.WriteString("\n  Tryck Enter for att avsluta...")
	}

	return b.String()
}

// renderSessionGraph creates an ASCII graph of WPM over time
func (m *Model) renderSessionGraph(currentWPM float64) string {
	sessions := make([]stats.SessionSummary, len(m.allStats.Sessions)+1)
	copy(sessions, m.allStats.Sessions)
	sessions[len(m.allStats.Sessions)] = stats.SessionSummary{WPM: currentWPM}

	if len(sessions) < 2 {
		return ""
	}

	return renderGraph(sessions)
}

// getPerformanceMessage returns a motivational message based on performance
func getPerformanceMessage(wpm, accuracy float64) string {
	switch {
	case wpm >= 80 && accuracy >= 98:
		return "Fantastisk! Snabb och exakt!"
	case wpm >= 80:
		return "Blixtrande snabbt! Jobba på noggrannheten."
	case accuracy >= 98:
		return "Noggrann som en schweizisk klocka!"
	case wpm >= 60 && accuracy >= 95:
		return "Riktigt bra! Håll det uppe!"
	case wpm >= 60:
		return "Bra tempo! Fokusera på att minska felen."
	case accuracy >= 95:
		return "Fin precision! Försök öka farten."
	case wpm >= 40 && accuracy >= 90:
		return "Godkant! Börja vaxla upp."
	case wpm < 30:
		return "Lugn och fin, du lär dig!"
	default:
		return "Övning ger färdighet. Fortsätt!"
	}
}
