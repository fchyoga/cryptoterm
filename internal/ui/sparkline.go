package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	sparkBlocks = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
)

// GenerateSparkline produces a compact Unicode sparkline from a series of floats.
func GenerateSparkline(data []float64, width int) string {
	if len(data) == 0 {
		return strings.Repeat(" ", width)
	}

	if len(data) == 1 {
		return strings.Repeat("━", width)
	}

	// Find min and max
	min := data[0]
	max := data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Flat line for stablecoins or constant data
	if max == min || max-min < 0.000000001 {
		flatStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
		return flatStyle.Render(strings.Repeat("━", width))
	}

	// Take up to `width` most recent samples
	samples := data
	if len(samples) > width {
		samples = samples[len(samples)-width:]
	}

	var sb strings.Builder
	diff := max - min
	numBlocks := float64(len(sparkBlocks) - 1)

	for _, v := range samples {
		normalized := (v - min) / diff
		idx := int(normalized * numBlocks)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkBlocks) {
			idx = len(sparkBlocks) - 1
		}
		sb.WriteRune(sparkBlocks[idx])
	}

	// Pad with spaces if fewer samples than width
	raw := sb.String()
	if len(samples) < width {
		pad := strings.Repeat(" ", width-len(samples))
		raw = pad + raw
	}

	// Color sparkline based on trend: compare last value to first value
	first := samples[0]
	last := samples[len(samples)-1]

	var style lipgloss.Style
	if last > first {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")) // Green
	} else if last < first {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")) // Red
	} else {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")) // Gray
	}

	return style.Render(raw)
}
