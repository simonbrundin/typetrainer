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
	correctStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	currentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	typedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	progressStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	borderStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

// introView renders the intro screen
func (m *Model) introView() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(borderStyle.Render("╔══════════════════════════════════════════╗\n"))
	b.WriteString(fmt.Sprintf("║          TYPE TRAINER - %-18s║\n", m.mode.String()))
	b.WriteString(borderStyle.Render("╠══════════════════════════════════════════╣\n"))
	b.WriteString("║  Skriv orden som visas.                 ║\n")
	b.WriteString("║  Tryck Ctrl+C för att avsluta.          ║\n")
	b.WriteString(borderStyle.Render("╚══════════════════════════════════════════╝\n"))
	b.WriteString("\n  Tryck Enter för att börja...")
	return b.String()
}

// sessionView renders the active typing session
func (m *Model) sessionView() string {
	var b strings.Builder

	s := m.session
	buf := s.Buffer
	text := buf.Text
	cursor := buf.Cursor

	b.WriteString("\n\n")

	// Convert to runes for proper Unicode handling
	runes := []rune(text)
	totalLen := len(runes)

	// Display settings
	const leftPadding = 3  // Spaces before the text starts
	const historyLen = 10  // How many typed chars to show (grayed out)

	// Calculate visible range
	// Show some history to the left (grayed out), then current and remaining
	startRune := cursor - historyLen
	if startRune < 0 {
		startRune = 0
	}
	endRune := cursor + DisplayLength - historyLen
	if endRune > totalLen {
		endRune = totalLen
	}

	// Left padding
	b.WriteString("  ")
	for i := 0; i < leftPadding; i++ {
		b.WriteString(" ")
	}

	// Show text with highlighting
	for i := startRune; i < endRune; i++ {
		r := runes[i]
		runeStr := string(r)

		if i < cursor {
			// Already typed - grayed out (but visible)
			b.WriteString(typedStyle.Render(runeStr))
		} else if i == cursor {
			// Current position - bright/bold
			b.WriteString(currentStyle.Render(runeStr))
		} else {
			// Remaining - dim
			b.WriteString(dimStyle.Render(runeStr))
		}
	}

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
		b.WriteString(dimStyle.Render("  │  Tab/Ctrl+Backspace = starta om"))
	}

	return b.String()
}

// summaryView renders the session summary
func (m *Model) summaryView() string {
	s := m.session
	wpm := s.WPM()
	rawWpm := s.RawWPM()
	accuracy := s.Accuracy()
	charsTyped := s.Buffer.Cursor

	msg := getPerformanceMessage(wpm, accuracy)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(borderStyle.Render("╔══════════════════════════════════════════╗\n"))
	b.WriteString(borderStyle.Render("║            SESSION COMPLETE!             ║\n"))
	b.WriteString(borderStyle.Render("╠══════════════════════════════════════════╣\n"))
	b.WriteString(fmt.Sprintf("║                                          ║\n"))
	b.WriteString(fmt.Sprintf("║   Tecken skrivna:     %12d        ║\n", charsTyped))
	b.WriteString(fmt.Sprintf("║   Totalt fel:          %12d        ║\n", s.TotalErrors))
	b.WriteString(fmt.Sprintf("║   Raw WPM:             %12.1f        ║\n", rawWpm))
	b.WriteString(fmt.Sprintf("║   WPM (korrekt):       %12.1f        ║\n", wpm))
	b.WriteString(fmt.Sprintf("║   Accuracy:            %12.1f%%        ║\n", accuracy))
	b.WriteString(fmt.Sprintf("║                                          ║\n"))
	b.WriteString(fmt.Sprintf("║   %s                         ║\n", msg))
	b.WriteString(fmt.Sprintf("║                                          ║\n"))
	b.WriteString(borderStyle.Render("╚══════════════════════════════════════════╝\n"))

	// Add progress graph if we have stats
	if m.allStats != nil {
		graph := m.renderSessionGraph(wpm)
		if graph != "" {
			b.WriteString("\n")
			b.WriteString(graph)
		}
	}

	if m.showContinue {
		b.WriteString(dimStyle.Render("  Enter = kör igen   |   Esc = avsluta"))
	} else {
		b.WriteString("\n  Tryck Enter för att avsluta...")
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
		return "Fantastisk! Du är snabb och exakt! 🎉"
	case wpm >= 80:
		return "Blixtrande snabbt! Jobba på noggrannheten."
	case accuracy >= 98:
		return "Noggrann som en schweizisk klocka! 💯"
	case wpm >= 60 && accuracy >= 95:
		return "Riktigt bra! Håll det uppe!"
	case wpm >= 60:
		return "Bra tempo! Fokusera på att minska felen."
	case accuracy >= 95:
		return "Fin precision! Försök öka farten."
	case wpm >= 40 && accuracy >= 90:
		return "Godkänt! Börja växla upp."
	case wpm < 30:
		return "Lugn och fin, du lär dig!"
	default:
		return "Övning ger färdighet. Fortsätt!"
	}
}
