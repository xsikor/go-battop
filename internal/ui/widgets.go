package ui

import (
	"fmt"
	"strings"
)

// Box drawing characters
const (
	// Single line box drawing
	BoxTopLeft     = "┌"
	BoxTopRight    = "┐"
	BoxBottomLeft  = "└"
	BoxBottomRight = "┘"
	BoxHorizontal  = "─"
	BoxVertical    = "│"
	BoxCross       = "┼"
	BoxTeeDown     = "┬"
	BoxTeeUp       = "┴"
	BoxTeeRight    = "├"
	BoxTeeLeft     = "┤"

	// Double line box drawing
	BoxDoubleHorizontal = "═"
	BoxDoubleVertical   = "║"

	// Block elements
	BlockFull   = "█"
	BlockLight  = "░"
	BlockMedium = "▒"
	BlockDark   = "▓"

	// Progress bar characters
	BarFull    = "█"
	BarEmpty   = "░"
	BarPartial = "▌"
)

// DrawBox draws a box with the given title
func DrawBox(width, height int, title string) []string {
	lines := make([]string, height)

	// Top line with title
	titleLen := len(title)
	if titleLen > width-4 {
		title = title[:width-4]
		titleLen = width - 4
	}

	padding := (width - titleLen - 2) / 2
	topLine := BoxTopLeft + strings.Repeat(BoxHorizontal, padding) + " " + title + " "
	topLine += strings.Repeat(BoxHorizontal, width-len(topLine)-1) + BoxTopRight
	lines[0] = topLine

	// Middle lines
	for i := 1; i < height-1; i++ {
		lines[i] = BoxVertical + strings.Repeat(" ", width-2) + BoxVertical
	}

	// Bottom line
	lines[height-1] = BoxBottomLeft + strings.Repeat(BoxHorizontal, width-2) + BoxBottomRight

	return lines
}

// CreateGradientBar creates a gradient progress bar
func CreateGradientBar(percent float64, width int) string {
	if width <= 0 {
		return ""
	}

	filled := int(percent * float64(width) / 100)
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	// Create gradient effect
	var bar strings.Builder

	for i := 0; i < width; i++ {
		if i < filled {
			// Full blocks for filled portion
			bar.WriteString(BarFull)
		} else if i == filled && percent*float64(width)/100 > float64(filled) {
			// Partial block for the transition
			bar.WriteString(BarPartial)
		} else {
			// Empty blocks for unfilled portion
			bar.WriteString(BarEmpty)
		}
	}

	return bar.String()
}

// FormatPercentage formats a percentage with color
func FormatPercentage(value float64, showSign bool) string {
	color := getPercentageColor(value)
	sign := ""
	if showSign && value > 0 {
		sign = "+"
	}
	return fmt.Sprintf("[%s]%s%.1f%%[-]", color, sign, value)
}

// getPercentageColor returns appropriate color for percentage
func getPercentageColor(percent float64) string {
	switch {
	case percent >= 80:
		return "green"
	case percent >= 60:
		return "yellow"
	case percent >= 40:
		return "orange"
	case percent >= 20:
		return "red"
	default:
		return "darkred"
	}
}

// CenterText centers text within a given width
func CenterText(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}

	padding := (width - textLen) / 2
	result := strings.Repeat(" ", padding) + text
	result += strings.Repeat(" ", width-len(result))

	return result
}

// TruncateText truncates text to fit within width, adding ellipsis if needed
func TruncateText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	if width <= 3 {
		return text[:width]
	}

	return text[:width-3] + "..."
}
