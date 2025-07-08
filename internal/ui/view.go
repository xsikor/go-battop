package ui

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xsikor/go-battop/internal/battery"
)

// ChartData holds time-series data for charts
type ChartData struct {
	timestamps []time.Time
	values     []float64
	maxSize    int
}

// NewChartData creates new chart data storage
func NewChartData(maxSize int) *ChartData {
	return &ChartData{
		timestamps: make([]time.Time, 0, maxSize),
		values:     make([]float64, 0, maxSize),
		maxSize:    maxSize,
	}
}

// Add adds a new data point
func (cd *ChartData) Add(value float64) {
	cd.timestamps = append(cd.timestamps, time.Now())
	cd.values = append(cd.values, value)

	// Remove old data if we exceed max size
	if len(cd.values) > cd.maxSize {
		cd.timestamps = cd.timestamps[1:]
		cd.values = cd.values[1:]
	}
}

// View represents a single battery view
type View struct {
	root        *tview.Flex
	infoText    *tview.TextView
	chargeGauge *tview.TextView
	powerGauge  *tview.TextView
	healthGauge *tview.TextView
	chartArea   *tview.TextView

	index      int
	config     Config
	lastUpdate time.Time

	// Charts
	voltageChart *Chart
	powerChart   *Chart
	chargeChart  *Chart
	chartSet     *ChartSet

	// Track chart dimensions
	chartWidth  int
	chartHeight int
}

// NewView creates a new battery view
func NewView(index int, config Config) *View {
	v := &View{
		index:       index,
		config:      config,
		infoText:    tview.NewTextView(),
		chargeGauge: tview.NewTextView(),
		powerGauge:  tview.NewTextView(),
		healthGauge: tview.NewTextView(),
		chartArea:   tview.NewTextView(),
		chartWidth:  80, // Default width
		chartHeight: 20, // Default height
	}

	// Create charts
	v.voltageChart = NewChart("Voltage", 120, "V", "yellow")
	v.powerChart = NewChart("Power", 120, "W", "green")
	v.chargeChart = NewChart("Charge", 120, "%", "cyan")

	// Create chart set
	v.chartSet = NewChartSet()
	v.chartSet.AddChart(v.voltageChart)
	v.chartSet.AddChart(v.powerChart)
	v.chartSet.AddChart(v.chargeChart)

	// Configure text views
	v.infoText.SetDynamicColors(true).SetBackgroundColor(tcell.ColorDefault)
	v.chargeGauge.SetDynamicColors(true).SetBackgroundColor(tcell.ColorDefault)
	v.powerGauge.SetDynamicColors(true).SetBackgroundColor(tcell.ColorDefault)
	v.healthGauge.SetDynamicColors(true).SetBackgroundColor(tcell.ColorDefault)

	// Initialize text views with placeholder content
	v.infoText.SetText("[gray]Loading battery information...[-]")
	v.chargeGauge.SetText(" [gray]Loading charge data...[-]")
	v.powerGauge.SetText(" [gray]Loading power data...[-]")
	v.healthGauge.SetText(" [gray]Loading health data...[-]")

	// Configure chart area
	v.chartArea.SetDynamicColors(true).
		SetBackgroundColor(tcell.ColorDefault)

	// Build layout
	v.buildLayout()

	return v
}

// buildLayout builds the view layout
func (v *View) buildLayout() {
	slog.Debug("Building view layout")

	// Main container (horizontal split)
	v.root = tview.NewFlex().SetDirection(tview.FlexColumn)

	// Left panel (info and gauges)
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add battery info directly (no frame for now to test)
	leftPanel.AddItem(v.infoText, 0, 2, false)

	// Add gauges directly (no frames for now to test)
	leftPanel.AddItem(v.chargeGauge, 1, 0, false)
	leftPanel.AddItem(v.powerGauge, 1, 0, false)
	leftPanel.AddItem(v.healthGauge, 1, 0, false)

	// Right panel (charts) - no frame to maximize space
	// Option 1: Use percentage-based layout (current implementation)
	// Left panel gets 20% of space, right gets 80%
	v.root.AddItem(leftPanel, 0, 1, false)  // 20% of space (1/5)
	v.root.AddItem(v.chartArea, 0, 4, true) // 80% of space (4/5)

	// Option 2: Fixed width for left panel (uncomment to use)
	// This gives consistent left panel size regardless of terminal width
	// v.root.AddItem(leftPanel, 40, 0, false)  // Fixed 40 chars width
	// v.root.AddItem(v.chartArea, 0, 1, true)  // Remaining space

	slog.Debug("Layout build complete", "leftProportion", 1, "rightProportion", 4)
}

// GetRoot returns the root UI element
func (v *View) GetRoot() tview.Primitive {
	return v.root
}

// Update updates the view with new battery information
func (v *View) Update(info *battery.Info) {
	v.lastUpdate = time.Now()
	slog.Debug("Updating view", "batteryIndex", v.index)

	// Update chart data
	v.voltageChart.AddValue(info.Voltage)

	// Convert power to human-readable units if needed
	power := info.ChargeRate
	if v.config != nil {
		// For chart display, use raw watts
		power = info.ChargeRate / 1000.0
	}
	v.powerChart.AddValue(power)

	v.chargeChart.AddValue(info.ChargePercent())

	// Update info text
	v.updateInfoText(info)

	// Update gauges
	v.updateGauges(info)

	// Update charts with current dimensions
	_, _, w, h := v.chartArea.GetInnerRect()
	if w > 0 && h > 0 {
		v.chartWidth = w
		v.chartHeight = h
	} else {
		// Use defaults if dimensions not available yet
		v.chartWidth = 80
		v.chartHeight = 20
	}
	v.updateCharts()
}

// updateInfoText updates the battery information display
func (v *View) updateInfoText(info *battery.Info) {
	var text strings.Builder

	// State with color
	stateColor := getStateColor(info.State)
	fmt.Fprintf(&text, "[%s:b]%s[-]\n", stateColor, info.State.String())

	fmt.Fprintf(&text, "[gray]--------------------------------[-]\n")

	// Manufacturer and model info
	if info.Manufacturer != "" {
		fmt.Fprintf(&text, "[cyan]Make:[-]      %s\n", info.Manufacturer)
	}
	if info.Model != "" {
		fmt.Fprintf(&text, "[cyan]Model:[-]     %s\n", info.Model)
	}

	// Technology and basic info
	fmt.Fprintf(&text, "[cyan]Type:[-]      %s\n", info.Technology)

	// Voltage information
	fmt.Fprintf(&text, "[cyan]Voltage:[-]   %s ", v.config.FormatVoltage(info.Voltage))
	fmt.Fprintf(&text, "[gray](design: %s)[-]\n", v.config.FormatVoltage(info.DesignVoltage))

	fmt.Fprintf(&text, "\n")

	// Capacity information
	fmt.Fprintf(&text, "[cyan]Current:[-]   %s\n", v.config.FormatEnergy(info.Current))
	fmt.Fprintf(&text, "[cyan]Full:[-]      %s ", v.config.FormatEnergy(info.Full))

	// Show battery health as percentage of design capacity
	health := info.Health()
	healthColor := getHealthColor(health)
	fmt.Fprintf(&text, "[gray]([%s]%.1f%%[-] health)[-]\n", healthColor, health)

	fmt.Fprintf(&text, "[cyan]Design:[-]    %s\n", v.config.FormatEnergy(info.Design))

	// Time remaining
	if info.State == battery.StateDischarging {
		if tte := info.TimeToEmpty(); tte > 0 {
			fmt.Fprintf(&text, "\n[orange]Time remaining: %s[-]\n", formatDuration(tte))
		}
	} else if info.State == battery.StateCharging {
		if ttf := info.TimeToFull(); ttf > 0 {
			fmt.Fprintf(&text, "\n[green]Time to full: %s[-]\n", formatDuration(ttf))
		}
	}

	// Cycle count if available
	if info.CycleCount > 0 {
		fmt.Fprintf(&text, "\n[cyan]Cycles:[-]    %d\n", info.CycleCount)
	}

	// Last update
	fmt.Fprintf(&text, "\n[gray]Updated: %s[-]", v.lastUpdate.Format("15:04:05"))

	finalText := text.String()
	slog.Debug("Updated info text", "length", len(finalText), "lines", strings.Count(finalText, "\n"))
	v.infoText.SetText(finalText)
}

// updateGauges updates the gauge displays
func (v *View) updateGauges(info *battery.Info) {
	// Charge gauge with gradient effect
	chargePercent := info.ChargePercent()
	chargeColor := getChargeColor(chargePercent)
	chargeBar := createGradientProgressBar(chargePercent, 20, chargeColor)
	chargeText := fmt.Sprintf(" %s [%s]%.1f%%[-]", chargeBar, chargeColor, chargePercent)
	v.chargeGauge.SetText(chargeText)
	slog.Debug("Updated charge gauge", "percent", chargePercent, "text", chargeText)

	// Power gauge with visual indicator
	var powerText string
	absPower := math.Abs(info.ChargeRate)
	if info.ChargeRate > 0 {
		// Charging
		powerText = fmt.Sprintf(" [green]>>> CHARGING[-] [white]%s[-]", v.config.FormatPower(absPower))
	} else if info.ChargeRate < 0 {
		// Discharging
		powerText = fmt.Sprintf(" [orange]<<< DISCHARGING[-] [white]%s[-]", v.config.FormatPower(absPower))
	} else {
		// Idle state
		powerText = fmt.Sprintf(" [gray]=== IDLE[-] [gray]%s[-]", v.config.FormatPower(0))
	}
	v.powerGauge.SetText(powerText)
	slog.Debug("Updated power gauge", "chargeRate", info.ChargeRate, "text", powerText)

	// Health gauge with gradient
	healthPercent := info.Health()
	healthColor := getHealthColor(healthPercent)
	healthBar := createGradientProgressBar(healthPercent, 20, healthColor)
	healthText := fmt.Sprintf(" %s [%s]%.1f%%[-]", healthBar, healthColor, healthPercent)
	v.healthGauge.SetText(healthText)
	slog.Debug("Updated health gauge", "percent", healthPercent, "text", healthText)
}

// updateCharts updates the chart display
func (v *View) updateCharts() {
	// Use the tracked dimensions
	width := v.chartWidth
	height := v.chartHeight

	// Don't update with invalid dimensions
	if width <= 0 || height <= 0 {
		slog.Debug("Skipping chart update - invalid dimensions", "width", width, "height", height)
		return
	}

	slog.Debug("Updating charts", "width", width, "height", height)

	// Add title
	var fullText strings.Builder
	title := " Real-time Monitoring "
	titleLen := len(title)
	if width > titleLen {
		padding := (width - titleLen) / 2
		titleLine := fmt.Sprintf("[white::b]%s%s%s[-]",
			strings.Repeat("─", padding),
			title,
			strings.Repeat("─", width-padding-titleLen))
		fullText.WriteString(titleLine)
		fullText.WriteString("\n")
	}

	// Update chart sizes (account for title)
	v.chartSet.SetSize(width, height-1)

	// Render and set the chart text
	chartText := v.chartSet.Render()
	if chartText == "" {
		slog.Warn("Chart render returned empty string")
	} else {
		lines := strings.Split(chartText, "\n")
		slog.Debug("Chart rendered", "lines", len(lines), "firstLine", lines[0])
		if len(lines) > 1 {
			slog.Debug("Chart second line", "line", lines[1])
		}
		fullText.WriteString(chartText)
	}

	v.chartArea.Clear()
	v.chartArea.SetText(fullText.String())
}

// Helper functions

func createGradientProgressBar(percent float64, width int, color string) string {
	filled := int(percent * float64(width) / 100)
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	var bar strings.Builder
	bar.WriteString(fmt.Sprintf("[%s]", color))

	// Use simple ASCII characters for compatibility
	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteString("=") // Filled
		} else {
			bar.WriteString("-") // Empty
		}
	}

	bar.WriteString("[-]")
	return bar.String()
}

func getChargeColor(percent float64) string {
	if percent >= 80 {
		return "green"
	} else if percent >= 50 {
		return "yellow"
	} else if percent >= 20 {
		return "orange"
	}
	return "red"
}

func getHealthColor(percent float64) string {
	if percent >= 80 {
		return "green"
	} else if percent >= 60 {
		return "yellow"
	}
	return "red"
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func getStateColor(state battery.State) string {
	switch state {
	case battery.StateCharging:
		return "green"
	case battery.StateDischarging:
		return "orange"
	case battery.StateFull:
		return "green"
	case battery.StateEmpty:
		return "red"
	case battery.StateNotCharging:
		return "yellow"
	default:
		return "white"
	}
}
