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
	if numPoints > 30 {
		numPoints = 30
	}
	startIdx := len(sessions) - numPoints

	// Extract WPM values
	wpmValues := extractWPMValues(sessions, startIdx, numPoints)
	avgWPM := calculateAverage(wpmValues)

	// Calculate min/max for scaling
	minWPM, maxWPM := findMinMax(wpmValues)
	minWPM, maxWPM = addPadding(minWPM, maxWPM)

	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("  WPM-utveckling (senaste 30) - Snitt: %.1f\n\n", avgWPM))

	// Draw grid
	drawGrid(&b, minWPM, maxWPM)

	// Calculate y positions for the line
	yPositions := calculateYPositions(wpmValues, minWPM, maxWPM)

	// Draw the line over the grid
	drawLine(&b, yPositions, numPoints)

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

// drawGrid draws the background grid
func drawGrid(b *strings.Builder, minWPM, maxWPM float64) {
	for row := graphHeight; row >= 0; row-- {
		yVal := minWPM + (maxWPM-minWPM)*float64(row)/float64(graphHeight)
		b.WriteString(fmt.Sprintf("  %4.0f |", yVal))

		for x := 0; x < graphWidth; x++ {
			b.WriteString(dimStyle.Render("·"))
		}
		b.WriteString("\n")
	}

	// Bottom axis
	b.WriteString("       +")
	b.WriteString(strings.Repeat("-", graphWidth))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("       %s%s%s\n", "första", strings.Repeat(" ", graphWidth-12), "senaste"))
}

// calculateYPositions converts WPM values to Y positions in the graph
func calculateYPositions(values []float64, minWPM, maxWPM float64) []int {
	numPoints := len(values)
	positions := make([]int, numPoints)

	for x := 0; x < numPoints && x < graphWidth; x++ {
		yFloat := (values[x] - minWPM) / (maxWPM - minWPM) * float64(graphHeight)
		pos := graphHeight - int(yFloat+0.5)
		if pos < 0 {
			pos = 0
		}
		if pos > graphHeight {
			pos = graphHeight
		}
		positions[x] = pos
	}
	return positions
}

// drawLine draws the WPM line over the grid
func drawLine(b *strings.Builder, yPositions []int, numPoints int) {
	b.WriteString("       |")
	for x := 0; x < graphWidth && x < numPoints; x++ {
		isPoint := isPeakOrValley(yPositions, x)

		if isPoint {
			b.WriteString(correctStyle.Render("█"))
		} else {
			b.WriteString(correctStyle.Render("▄"))
		}
	}
	b.WriteString("\n")
}

// isPeakOrValley determines if a point should be drawn as a full block
func isPeakOrValley(positions []int, x int) bool {
	if x <= 0 || x >= len(positions)-1 || x >= graphWidth-1 {
		return true
	}
	prevY := positions[x-1]
	nextY := positions[x+1]
	return positions[x] != prevY && positions[x] != nextY
}

// lastValue returns the last element or 0 if empty
func lastValue(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return values[len(values)-1]
}
