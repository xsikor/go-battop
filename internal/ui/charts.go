package ui

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"
)

// Chart represents a time-series chart
type Chart struct {
	title     string
	data      *ChartData
	width     int
	height    int
	minValue  float64
	maxValue  float64
	autoScale bool
	unit      string
	color     string
}

// NewChart creates a new chart
func NewChart(title string, maxDataPoints int, unit string, color string) *Chart {
	return &Chart{
		title:     title,
		data:      NewChartData(maxDataPoints),
		autoScale: true,
		unit:      unit,
		color:     color,
	}
}

// SetSize sets the chart dimensions
func (c *Chart) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// SetScale sets manual scale for the chart
func (c *Chart) SetScale(min, max float64) {
	c.minValue = min
	c.maxValue = max
	c.autoScale = false
}

// AddValue adds a new value to the chart
func (c *Chart) AddValue(value float64) {
	c.data.Add(value)
}

// Render renders the chart as a string
func (c *Chart) Render() string {
	slog.Debug("Chart.Render", "title", c.title, "width", c.width, "height", c.height, "dataPoints", len(c.data.values))

	if c.width <= 0 || c.height <= 0 {
		return " [gray]Initializing...[-]"
	}

	if len(c.data.values) == 0 {
		return c.renderEmptyChart()
	}

	var result strings.Builder
	min, max := c.calculateBounds()

	c.renderTitle(&result)
	c.renderChartBody(&result, min, max)
	c.renderXAxis(&result)
	result.WriteString(c.createTimeLabels())

	return result.String()
}

// renderTitle renders the chart title with decorative borders
func (c *Chart) renderTitle(result *strings.Builder) {
	titleStr := c.prepareTitleString()
	leftPad, rightPad := c.calculateTitlePadding(titleStr)

	if leftPad > 0 {
		result.WriteString(strings.Repeat("─", leftPad))
	}
	result.WriteString(fmt.Sprintf("[%s:b]%s[-]", c.color, titleStr))
	if rightPad > 0 {
		result.WriteString(strings.Repeat("─", rightPad))
	}
	result.WriteString("\n")
}

// prepareTitleString prepares the title string, truncating if necessary
func (c *Chart) prepareTitleString() string {
	titleStr := fmt.Sprintf(" %s ", c.title)
	titleLen := len(titleStr)

	if c.width < titleLen {
		// Truncate title if too long
		return c.title[:c.width-2] + " "
	}
	return titleStr
}

// calculateTitlePadding calculates left and right padding for centered title
func (c *Chart) calculateTitlePadding(titleStr string) (int, int) {
	titleLen := len(titleStr)
	totalPadding := c.width - titleLen

	if totalPadding <= 0 {
		return 0, 0
	}

	leftPad := totalPadding / 2
	rightPad := totalPadding - leftPad
	return leftPad, rightPad
}

// renderChartBody renders the Y-axis labels and chart content
func (c *Chart) renderChartBody(result *strings.Builder, min, max float64) {
	chartHeight := c.calculateChartHeight()
	grid := c.createGrid(min, max, chartHeight)

	for i := 0; i < chartHeight; i++ {
		yValue := c.calculateYValue(i, chartHeight, min, max)
		label := c.formatValue(yValue)

		result.WriteString(fmt.Sprintf("[gray]%8s ┤[-] ", label))
		result.WriteString(grid[i])
		result.WriteString("\n")
	}
}

// calculateChartHeight calculates the effective chart height
func (c *Chart) calculateChartHeight() int {
	chartHeight := c.height - ChartHeightReserve
	if chartHeight < MinChartHeight {
		return MinChartHeight
	}
	return chartHeight
}

// calculateYValue calculates the Y-axis value for a given row
func (c *Chart) calculateYValue(row, totalRows int, min, max float64) float64 {
	if totalRows <= 1 {
		return max
	}
	return max - (float64(row)/float64(totalRows-1))*(max-min)
}

// renderXAxis renders the X-axis decoration
func (c *Chart) renderXAxis(result *strings.Builder) {
	result.WriteString(fmt.Sprintf("[gray]%8s └", ""))
	result.WriteString(strings.Repeat("─", c.width-YAxisLabelWidth))
	result.WriteString("[-]\n")
}

// calculateBounds calculates the min and max values for the chart
func (c *Chart) calculateBounds() (float64, float64) {
	if !c.autoScale {
		return c.minValue, c.maxValue
	}

	if len(c.data.values) == 0 {
		return 0, 1
	}

	min, max := c.data.values[0], c.data.values[0]
	for _, v := range c.data.values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Add some padding
	range_ := max - min
	if range_ < 0.001 {
		// If values are too close, add artificial range
		min = min - 0.5
		max = max + 0.5
		return min, max
	}

	padding := range_ * 0.1
	min = min - padding
	max = max + padding

	return min, max
}

// createGrid creates the chart grid with data points
func (c *Chart) createGrid(min, max float64, height int) []string {
	chartWidth := c.calculateEffectiveChartWidth()
	grid := c.initializeEmptyGrid(height, chartWidth)

	if len(c.data.values) == 0 {
		return grid
	}

	c.plotDataPoints(grid, min, max, height, chartWidth)
	c.applyColorToGrid(grid)

	return grid
}

// calculateEffectiveChartWidth calculates the chart width minus Y-axis labels
func (c *Chart) calculateEffectiveChartWidth() int {
	return c.width - YAxisLabelWidth
}

// initializeEmptyGrid creates an empty grid filled with spaces
func (c *Chart) initializeEmptyGrid(height, width int) []string {
	grid := make([]string, height)
	for i := 0; i < height; i++ {
		grid[i] = strings.Repeat(" ", width)
	}
	return grid
}

// plotDataPoints plots all data points on the grid
func (c *Chart) plotDataPoints(grid []string, min, max float64, height, chartWidth int) {
	startIdx, endIdx := c.calculateVisibleDataRange(chartWidth)

	for i := startIdx; i < endIdx; i++ {
		x := i - startIdx
		if x >= chartWidth {
			break
		}

		c.plotSinglePoint(grid, i, x, min, max, height, chartWidth, startIdx)
	}
}

// calculateVisibleDataRange determines which data points are visible
func (c *Chart) calculateVisibleDataRange(chartWidth int) (int, int) {
	dataPoints := len(c.data.values)
	startIdx := 0
	if dataPoints > chartWidth {
		startIdx = dataPoints - chartWidth
	}
	return startIdx, dataPoints
}

// plotSinglePoint plots a single data point and connects it to the previous point
func (c *Chart) plotSinglePoint(grid []string, dataIdx, x int, min, max float64, height, chartWidth, startIdx int) {
	value := c.data.values[dataIdx]
	y := c.valueToY(value, min, max, height)

	// Plot the point
	if y >= 0 && y < height {
		c.setGridPoint(grid, x, y, dataIdx, height, min, max)
	}

	// Connect to previous point
	if dataIdx > startIdx {
		prevValue := c.data.values[dataIdx-1]
		prevY := c.valueToY(prevValue, min, max, height)
		c.drawVerticalLine(grid, x, prevY, y, chartWidth, height)
	}
}

// setGridPoint sets a single point on the grid with the appropriate character
func (c *Chart) setGridPoint(grid []string, x, y, dataIdx, height int, min, max float64) {
	line := []rune(grid[y])
	if x < len(line) {
		char := c.getPlotChar(dataIdx, y, height, min, max)
		line[x] = char
		grid[y] = string(line)
	}
}

// applyColorToGrid applies the chart color to all grid lines
func (c *Chart) applyColorToGrid(grid []string) {
	for i := range grid {
		grid[i] = fmt.Sprintf("[%s]%s[-]", c.color, grid[i])
	}
}

// valueToY converts a value to Y coordinate
func (c *Chart) valueToY(value, min, max float64, height int) int {
	if max <= min {
		return height / 2
	}
	normalized := (value - min) / (max - min)
	y := int(float64(height-1) * (1 - normalized))
	if y < 0 {
		y = 0
	}
	if y >= height {
		y = height - 1
	}
	return y
}

// getPlotChar determines which character to use for plotting
func (c *Chart) getPlotChar(dataIdx, y, height int, min, max float64) rune {
	// For the most recent data point, use a different character
	if dataIdx == len(c.data.values)-1 {
		return '*' // Current value
	}

	// Check if this is a peak or valley
	if dataIdx > 0 && dataIdx < len(c.data.values)-1 {
		prev := c.data.values[dataIdx-1]
		next := c.data.values[dataIdx+1]

		prevY := c.valueToY(prev, min, max, height)
		nextY := c.valueToY(next, min, max, height)

		if y < prevY && y < nextY {
			return '/' // Peak
		}
		if y > prevY && y > nextY {
			return '\\' // Valley
		}
	}

	return 'o' // Regular point
}

// drawVerticalLine draws a vertical line between two points
func (c *Chart) drawVerticalLine(grid []string, x, y1, y2 int, width, height int) {
	if x >= width || x < 0 {
		return
	}

	start := y1
	end := y2
	if start > end {
		start, end = end, start
	}

	for y := start; y <= end; y++ {
		if y >= 0 && y < height && y != y1 && y != y2 {
			line := []rune(grid[y])
			if x < len(line) && line[x] == ' ' {
				line[x] = '│'
			}
			grid[y] = string(line)
		}
	}
}

// formatValue formats a value for display
func (c *Chart) formatValue(value float64) string {
	// Determine appropriate precision based on value magnitude
	absValue := math.Abs(value)

	switch {
	case absValue >= 1000:
		return fmt.Sprintf("%.0f%s", value, c.unit)
	case absValue >= 10:
		return fmt.Sprintf("%.1f%s", value, c.unit)
	case absValue >= 1:
		return fmt.Sprintf("%.2f%s", value, c.unit)
	default:
		return fmt.Sprintf("%.3f%s", value, c.unit)
	}
}

// renderEmptyChart renders an empty chart with axes
func (c *Chart) renderEmptyChart() string {
	if c.width <= 0 || c.height <= 0 {
		return ""
	}

	var result strings.Builder

	// Title with decoration
	titleStr := fmt.Sprintf(" %s ", c.title)
	titleLen := len(titleStr)

	if c.width < titleLen {
		titleStr = c.title[:c.width-2] + " "
		titleLen = len(titleStr)
	}

	sidePadding := (c.width - titleLen) / 2
	if sidePadding < 0 {
		sidePadding = 0
	}

	remainingPadding := c.width - titleLen - sidePadding
	if remainingPadding < 0 {
		remainingPadding = 0
	}

	if sidePadding > 0 {
		result.WriteString(strings.Repeat("─", sidePadding))
	}
	result.WriteString(fmt.Sprintf("[%s:b]%s[-]", c.color, titleStr))
	if remainingPadding > 0 {
		result.WriteString(strings.Repeat("─", remainingPadding))
	}
	result.WriteString("\n")

	// Y-axis labels and empty chart area
	chartHeight := c.height - ChartHeightReserve
	if chartHeight < 2 {
		chartHeight = 2
	}

	// Draw Y-axis with default scale
	minVal := 0.0
	maxVal := 100.0
	if c.unit == "V" {
		minVal = 0.0
		maxVal = 20.0
	}
	if c.unit == "W" {
		minVal = -20.0
		maxVal = 20.0
	}

	for i := 0; i < chartHeight; i++ {
		yValue := maxVal - (float64(i)/float64(chartHeight-1))*(maxVal-minVal)
		label := c.formatValue(yValue)
		result.WriteString(fmt.Sprintf("[gray]%8s ┤[-] ", label))

		// Empty chart line
		result.WriteString(fmt.Sprintf("[gray]%s[-]\n", strings.Repeat("·", c.width-11)))
	}

	// X-axis
	result.WriteString(fmt.Sprintf("[gray]%8s └", ""))
	result.WriteString(strings.Repeat("─", c.width-11))
	result.WriteString("[-]\n")

	// Time labels placeholder
	result.WriteString(fmt.Sprintf("[gray]%8s   Waiting for data...[-]", ""))

	return result.String()
}

// createTimeLabels creates time labels for x-axis
func (c *Chart) createTimeLabels() string {
	if len(c.data.timestamps) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("[gray]%8s   ", ""))

	chartWidth := c.width - 11

	// Show time labels at start, middle, and end
	if len(c.data.timestamps) > 0 {
		// Calculate time range
		startTime := c.data.timestamps[0]
		endTime := c.data.timestamps[len(c.data.timestamps)-1]
		duration := endTime.Sub(startTime)

		// Start time
		result.WriteString(fmt.Sprintf("[gray]%s", startTime.Format("15:04:05")))

		// Calculate spacing
		labelWidth := 8
		spacing := chartWidth - (3 * labelWidth)
		if spacing > 0 && len(c.data.timestamps) > 1 {
			// Middle section with duration info
			midSpacing := spacing / 2
			if midSpacing > 4 {
				result.WriteString(strings.Repeat(" ", midSpacing-4))

				// Show duration in the middle
				durationStr := fmt.Sprintf("[cyan](%s)[-]", formatChartDuration(duration))
				result.WriteString(durationStr)

				// End time
				remainingSpace := spacing - midSpacing - 4
				if remainingSpace > 0 {
					result.WriteString(strings.Repeat(" ", remainingSpace))
				}
				result.WriteString(fmt.Sprintf("[gray]%s", endTime.Format("15:04:05")))
			} else {
				// Not enough space for duration, just add spacing
				result.WriteString(strings.Repeat(" ", spacing))
				result.WriteString(fmt.Sprintf("[gray]%s", endTime.Format("15:04:05")))
			}
		}
	}

	result.WriteString("[-]")
	return result.String()
}

// formatChartDuration formats duration for chart display
func formatChartDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

// ChartSet manages multiple charts
type ChartSet struct {
	charts []*Chart
	width  int
	height int
}

// NewChartSet creates a new chart set
func NewChartSet() *ChartSet {
	return &ChartSet{
		charts: make([]*Chart, 0),
	}
}

// AddChart adds a chart to the set
func (cs *ChartSet) AddChart(chart *Chart) {
	cs.charts = append(cs.charts, chart)
}

// SetSize sets the size for all charts
func (cs *ChartSet) SetSize(width, height int) {
	cs.width = width
	cs.height = height

	// Distribute height among charts
	if len(cs.charts) > 0 {
		chartHeight := height / len(cs.charts)
		slog.Debug("ChartSet SetSize", "width", width, "height", height, "chartCount", len(cs.charts), "chartHeight", chartHeight)
		for _, chart := range cs.charts {
			chart.SetSize(width, chartHeight)
		}
	}
}

// Render renders all charts
func (cs *ChartSet) Render() string {
	var result strings.Builder

	for i, chart := range cs.charts {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(chart.Render())
	}

	return result.String()
}
