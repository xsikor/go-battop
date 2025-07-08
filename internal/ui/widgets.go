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

// ProgressBarStyle defines the style of progress bar
type ProgressBarStyle struct {
	Full    string
	Empty   string
	Partial string
}

// Progress bar styles
var (
	ProgressBarStyleUnicode = ProgressBarStyle{
		Full:    BarFull,
		Empty:   BarEmpty,
		Partial: BarPartial,
	}
	ProgressBarStyleASCII = ProgressBarStyle{
		Full:    "=",
		Empty:   "-",
		Partial: "", // ASCII style doesn't use partial
	}
)

// CreateProgressBar creates a progress bar with customizable style
func CreateProgressBar(percent float64, width int, style ProgressBarStyle) string {
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

	var bar strings.Builder

	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteString(style.Full)
			continue
		}
		// Use partial block if available and appropriate
		if style.Partial != "" && i == filled && percent*float64(width)/100 > float64(filled) {
			bar.WriteString(style.Partial)
			continue
		}
		bar.WriteString(style.Empty)
	}

	return bar.String()
}

// CreateGradientBar creates a Unicode gradient progress bar (compatibility wrapper)
func CreateGradientBar(percent float64, width int) string {
	return CreateProgressBar(percent, width, ProgressBarStyleUnicode)
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

// ColorThresholds defines color thresholds for percentage values
type ColorThresholds struct {
	Excellent float64
	Good      float64
	Warning   float64
	Critical  float64
}

// Default color threshold presets
var (
	ColorThresholdsDefault = ColorThresholds{
		Excellent: 80,
		Good:      50,
		Warning:   20,
		Critical:  0,
	}

	ColorThresholdsHealth = ColorThresholds{
		Excellent: 80,
		Good:      60,
		Warning:   40,
		Critical:  0,
	}
)

// GetColorByThreshold returns appropriate color based on percentage and thresholds
func GetColorByThreshold(percent float64, thresholds ColorThresholds) string {
	if percent >= thresholds.Excellent {
		return "green"
	}
	if percent >= thresholds.Good {
		return "yellow"
	}
	if percent >= thresholds.Warning {
		return "orange"
	}
	return "red"
}

// getPercentageColor returns appropriate color for percentage (compatibility wrapper)
func getPercentageColor(percent float64) string {
	// Use a more granular threshold for general percentages
	if percent >= 80 {
		return "green"
	}
	if percent >= 60 {
		return "yellow"
	}
	if percent >= 40 {
		return "orange"
	}
	if percent >= 20 {
		return "red"
	}
	return "darkred"
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
