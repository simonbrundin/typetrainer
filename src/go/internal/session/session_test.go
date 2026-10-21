package session

import (
	"strings"
	"testing"
	"time"

	"typetrainer/internal/stats"
	"typetrainer/internal/wordlist"

	"github.com/charmbracelet/bubbletea"
)

// newTestModel bygger en modell med ett kort ord att skriva
func newTestModel(text string) *Model {
	cfg := stats.DefaultConfig()
	words := []*wordlist.Word{{Text: text, Frequency: 1}}

	return &Model{
		session: newSession(words, ModeAdaptive, &cfg),
		config:  &cfg,
		mode:    ModeAdaptive,
	}
}

func typeRune(m *Model, r rune) {
	m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
}

func TestModel_ErrorIsCountedAfterRestart(t *testing.T) {
	m := newTestModel("hej")

	// Skriv fel
	typeRune(m, 'x')
	if !m.session.HasSequenceErrors {
		t.Fatal("expected HasSequenceErrors after wrong keypress")
	}
	if m.session.TotalErrors != 1 {
		t.Errorf("TotalErrors = %d, want 1", m.session.TotalErrors)
	}

	// Starta om ordet
	m.handleKeyPress(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.session.HasSequenceErrors {
		t.Error("HasSequenceErrors should be cleared after restart")
	}

	// Skriv rätt
	typeRune(m, 'h')

	if len(m.session.Keystrokes) != 2 {
		t.Fatalf("keystrokes = %d, want 2 (felet ska behållas)", len(m.session.Keystrokes))
	}

	accuracy := m.session.Accuracy()
	if accuracy >= 100 {
		t.Errorf("Accuracy = %.1f, want < 100 (ett fel registrerades)", accuracy)
	}
}

func TestModel_RestartKeepsBigramContext(t *testing.T) {
	m := newTestModel("hej")

	// Skriv "h" och sedan fel på "e"
	typeRune(m, 'h')
	typeRune(m, 'x')
	m.handleKeyPress(tea.KeyMsg{Type: tea.KeyBackspace})

	// Efter omstart ska föregående tecken inte räknas som föregående
	if prev := m.getPreviousRune(); prev != 0 {
		t.Errorf("getPreviousRune() = %q, want 0 after restart", string(prev))
	}

	// Skriv "h" igen - nu ska föregående tecken vara 'h'
	typeRune(m, 'h')
	if prev := m.getPreviousRune(); prev != 'h' {
		t.Errorf("getPreviousRune() = %q, want 'h'", string(prev))
	}
}

func TestSession_WPMAndRawWPMAreConsistent(t *testing.T) {
	m := newTestModel("hej")

	typeRune(m, 'h')
	typeRune(m, 'e')

	// Ett fel och en omstart
	typeRune(m, 'x')
	m.handleKeyPress(tea.KeyMsg{Type: tea.KeyBackspace})
	typeRune(m, 'h')
	typeRune(m, 'e')

	raw := m.session.RawWPM()
	net := m.session.WPM()

	if raw <= 0 {
		t.Errorf("RawWPM = %.1f, want > 0", raw)
	}
	// Raw ska räkna alla tangenttryck, net bara korrekta tecken i texten
	if net <= 0 {
		t.Errorf("WPM = %.1f, want > 0", net)
	}
	if raw < net {
		t.Errorf("RawWPM (%.1f) should not be lower than WPM (%.1f)", raw, net)
	}
}

func TestSession_FreezeKeepsValuesStable(t *testing.T) {
	m := newTestModel("hej")

	typeRune(m, 'h')
	typeRune(m, 'e')

	m.session.CompletedAt = time.Now()
	m.session.Freeze()

	frozen := m.session.Result()
	time.Sleep(20 * time.Millisecond)

	if again := m.session.Result(); again.WPM != frozen.WPM {
		t.Errorf("WPM changed after freeze: %.2f -> %.2f", frozen.WPM, again.WPM)
	}
	if again := m.session.Result(); again.DurationMs != frozen.DurationMs {
		t.Errorf("DurationMs changed after freeze: %d -> %d", frozen.DurationMs, again.DurationMs)
	}
}

func TestSession_WordsCompleted(t *testing.T) {
	m := newTestModel("hej")

	if got := m.session.WordsCompleted(); got != 0 {
		t.Errorf("WordsCompleted() = %d, want 0", got)
	}

	typeRune(m, 'h')
	typeRune(m, 'e')
	typeRune(m, 'j')

	if got := m.session.WordsCompleted(); got != 1 {
		t.Errorf("WordsCompleted() = %d, want 1", got)
	}
}

func TestRenderGraph_DrawsPointsAtDifferentRows(t *testing.T) {
	sessions := []stats.SessionSummary{
		{WPM: 10}, {WPM: 90}, {WPM: 30}, {WPM: 70},
	}

	graph := renderGraph(sessions)
	if graph == "" {
		t.Fatal("renderGraph returned empty string")
	}

	// Varje rad i grafen ska innehålla minst ett "█" om värdet är markerat
	lines := 0
	for _, line := range strings.Split(graph, "\n") {
		if strings.Contains(line, "█") {
			lines++
		}
	}
	if lines < 2 {
		t.Errorf("expected points on several rows, got %d:\n%s", lines, graph)
	}
}
