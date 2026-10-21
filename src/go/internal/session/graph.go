package session

import (
	"fmt"
	"strings"

	"typetrainer/internal/stats"
)

// Graph rendering functions for session performance visualization

// Graph dimensions
const (
	graphWidth  = 40
	graphHeight = 5
)

// renderGraph draws an ASCII graph of session WPM values
func renderGraph(sessions []stats.SessionSummary) string {
	numPoints := len(sessions)
	if numPoints > graphWidth {
		numPoints = graphWidth
	}
	if numPoints < 1 {
		return ""
	}
	startIdx := len(sessions) - numPoints

	// Extract WPM values
	wpmValues := extractWPMValues(sessions, startIdx, numPoints)
	avgWPM := calculateAverage(wpmValues)
	minWPM, maxWPM := findMinMax(wpmValues)
	minWPM, maxWPM = addPadding(minWPM, maxWPM)

	// Placera punkterna i rutnätet och rita linjen
	rows := positionPoints(wpmValues, minWPM, maxWPM)
	grid := drawGrid(rows)

	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("  WPM-utveckling (senaste %d) - Snitt: %.1f\n\n", numPoints, avgWPM))

	// Rita rader med etikett (överst = högst värde)
	for i := 0; i <= graphHeight; i++ {
		yVal := minWPM + (maxWPM-minWPM)*float64(graphHeight-i)/float64(graphHeight)
		b.WriteString(fmt.Sprintf("  %4.0f |", yVal))
		b.WriteString(grid[i])
		b.WriteString("\n")
	}

	// Bottom axis
	b.WriteString("       +")
	b.WriteString(strings.Repeat("-", graphWidth))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("       %s%s%s\n", "första", strings.Repeat(" ", graphWidth-12), "senaste"))

	// Legend
	b.WriteString(fmt.Sprintf("\n  %s Senaste: %.1f   %s Snitt: %.1f\n",
		correctStyle.Render("█"), lastValue(wpmValues),
		dimStyle.Render("---"), avgWPM))

	return b.String()
}

// extractWPMValues extracts WPM values from sessions
func extractWPMValues(sessions []stats.SessionSummary, startIdx, numPoints int) []float64 {
	values := make([]float64, numPoints)
	for i := 0; i < numPoints; i++ {
		values[i] = sessions[startIdx+i].WPM
	}
	return values
}

// calculateAverage computes the average of a slice
func calculateAverage(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// findMinMax finds the minimum and maximum values
func findMinMax(values []float64) (min, max float64) {
	min = values[0]
	max = values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// addPadding adds padding to the range for better visualization
func addPadding(min, max float64) (float64, float64) {
	padding := (max - min) * 0.1
	min -= padding
	max += padding
	if max-min < 10 {
		min -= 10
		max += 10
	}
	return min, max
}

// positionPoints maps each value to a row in the grid (0 = lägst, graphHeight = högst)
func positionPoints(values []float64, minWPM, maxWPM float64) []int {
	positions := make([]int, len(values))

	for i, v := range values {
		frac := 0.0
		if maxWPM > minWPM {
			frac = (v - minWPM) / (maxWPM - minWPM)
		}
		pos := int(frac*graphHeight + 0.5)
		if pos < 0 {
			pos = 0
		}
		if pos > graphHeight {
			pos = graphHeight
		}
		positions[i] = pos
	}
	return positions
}

// drawGrid renders the points and the line between them.
// Returnerar en rad per grafhöjd, med den översta raden först.
func drawGrid(positions []int) []string {
	cells := make([][]rune, graphHeight+1)
	for i := range cells {
		cells[i] = []rune(strings.Repeat("·", graphWidth))
	}

	for x := 0; x < len(positions) && x < graphWidth; x++ {
		row := positions[x]

		// Rita den vertikala delen av linjen mellan föregående och aktuell punkt
		if x > 0 {
			prev := positions[x-1]
			if prev != row {
				step := 1
				if prev > row {
					step = -1
				}
				for r := prev + step; r != row; r += step {
					cells[r][x-1] = '│'
				}
			}
		}

		cells[row][x] = '█'
	}

	rows := make([]string, graphHeight+1)
	for i := 0; i <= graphHeight; i++ {
		rows[i] = string(cells[graphHeight-i])
	}
	return rows
}

// lastValue returns the last element or 0 if empty
func lastValue(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return values[len(values)-1]
}
