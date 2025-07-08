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
		chartWidth:  DefaultChartWidth,
		chartHeight: DefaultChartHeight,
	}

	// Create charts
	v.voltageChart = NewChart("Voltage", MaxChartDataPoints, "V", "yellow")
	v.powerChart = NewChart("Power", MaxChartDataPoints, "W", "green")
	v.chargeChart = NewChart("Charge", MaxChartDataPoints, "%", "cyan")

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
	if w <= 0 || h <= 0 {
		// Use defaults if dimensions not available yet
		v.chartWidth = DefaultChartWidth
		v.chartHeight = DefaultChartHeight
		v.updateCharts()
		return
	}
	v.chartWidth = w
	v.chartHeight = h
	v.updateCharts()
}

// updateInfoText updates the battery information display
func (v *View) updateInfoText(info *battery.Info) {
	var text strings.Builder

	// Build each section
	v.addBatteryState(&text, info)
	v.addSeparator(&text)
	v.addBatteryIdentity(&text, info)
	v.addBatteryVoltage(&text, info)
	v.addBatteryCapacity(&text, info)
	v.addBatteryTimeRemaining(&text, info)
	v.addBatteryCycles(&text, info)
	v.addUpdateTimestamp(&text)

	finalText := text.String()
	slog.Debug("Updated info text", "length", len(finalText), "lines", strings.Count(finalText, "\n"))
	v.infoText.SetText(finalText)
}

// addBatteryState adds the battery state line
func (v *View) addBatteryState(text *strings.Builder, info *battery.Info) {
	stateColor := getStateColor(info.State)
	fmt.Fprintf(text, "[%s:b]%s[-]\n", stateColor, info.State.String())
}

// addSeparator adds a visual separator line
func (v *View) addSeparator(text *strings.Builder) {
	fmt.Fprintf(text, "[gray]--------------------------------[-]\n")
}

// addBatteryIdentity adds manufacturer, model, and type information
func (v *View) addBatteryIdentity(text *strings.Builder, info *battery.Info) {
	if info.Manufacturer != "" {
		fmt.Fprintf(text, "[cyan]Make:[-]      %s\n", info.Manufacturer)
	}
	if info.Model != "" {
		fmt.Fprintf(text, "[cyan]Model:[-]     %s\n", info.Model)
	}
	fmt.Fprintf(text, "[cyan]Type:[-]      %s\n", info.Technology)
}

// addBatteryVoltage adds voltage information
func (v *View) addBatteryVoltage(text *strings.Builder, info *battery.Info) {
	fmt.Fprintf(text, "[cyan]Voltage:[-]   %s ", v.config.FormatVoltage(info.Voltage))
	fmt.Fprintf(text, "[gray](design: %s)[-]\n\n", v.config.FormatVoltage(info.DesignVoltage))
}

// addBatteryCapacity adds capacity and health information
func (v *View) addBatteryCapacity(text *strings.Builder, info *battery.Info) {
	fmt.Fprintf(text, "[cyan]Current:[-]   %s\n", v.config.FormatEnergy(info.Current))
	fmt.Fprintf(text, "[cyan]Full:[-]      %s ", v.config.FormatEnergy(info.Full))

	// Show battery health as percentage of design capacity
	health := info.Health()
	healthColor := getHealthColor(health)
	fmt.Fprintf(text, "[gray]([%s]%.1f%%[-] health)[-]\n", healthColor, health)

	fmt.Fprintf(text, "[cyan]Design:[-]    %s\n", v.config.FormatEnergy(info.Design))
}

// addBatteryTimeRemaining adds time to empty/full information
func (v *View) addBatteryTimeRemaining(text *strings.Builder, info *battery.Info) {
	if info.State == battery.StateDischarging {
		if tte := info.TimeToEmpty(); tte > 0 {
			fmt.Fprintf(text, "\n[orange]Time remaining: %s[-]\n", formatDuration(tte))
		}
	}
	if info.State == battery.StateCharging {
		if ttf := info.TimeToFull(); ttf > 0 {
			fmt.Fprintf(text, "\n[green]Time to full: %s[-]\n", formatDuration(ttf))
		}
	}
}

// addBatteryCycles adds cycle count if available
func (v *View) addBatteryCycles(text *strings.Builder, info *battery.Info) {
	if info.CycleCount > 0 {
		fmt.Fprintf(text, "\n[cyan]Cycles:[-]    %d\n", info.CycleCount)
	}
}

// addUpdateTimestamp adds the last update timestamp
func (v *View) addUpdateTimestamp(text *strings.Builder) {
	fmt.Fprintf(text, "\n[gray]Updated: %s[-]", v.lastUpdate.Format(TimeFormat))
}

// updateGauges updates the gauge displays
func (v *View) updateGauges(info *battery.Info) {
	v.updateChargeGauge(info)
	v.updatePowerGauge(info)
	v.updateHealthGauge(info)
}

// updateChargeGauge updates the charge gauge display
func (v *View) updateChargeGauge(info *battery.Info) {
	chargePercent := info.ChargePercent()
	chargeColor := getChargeColor(chargePercent)
	chargeBar := CreateProgressBar(chargePercent, ProgressBarWidth, ProgressBarStyleASCII)
	chargeText := fmt.Sprintf(" [%s]%s[-] [%s]%.1f%%[-]", chargeColor, chargeBar, chargeColor, chargePercent)
	v.chargeGauge.SetText(chargeText)
	slog.Debug("Updated charge gauge", "percent", chargePercent, "text", chargeText)
}

// updatePowerGauge updates the power gauge display
func (v *View) updatePowerGauge(info *battery.Info) {
	var powerText string
	absPower := math.Abs(info.ChargeRate)

	// Idle state
	if info.ChargeRate == 0 {
		powerText = fmt.Sprintf(" [gray]=== IDLE[-] [gray]%s[-]", v.config.FormatPower(0))
		v.powerGauge.SetText(powerText)
		slog.Debug("Updated power gauge", "chargeRate", info.ChargeRate, "text", powerText)
		return
	}

	// Charging
	if info.ChargeRate > 0 {
		powerText = fmt.Sprintf(" [green]>>> CHARGING[-] [white]%s[-]", v.config.FormatPower(absPower))
		v.powerGauge.SetText(powerText)
		slog.Debug("Updated power gauge", "chargeRate", info.ChargeRate, "text", powerText)
		return
	}

	// Discharging
	powerText = fmt.Sprintf(" [orange]<<< DISCHARGING[-] [white]%s[-]", v.config.FormatPower(absPower))
	v.powerGauge.SetText(powerText)
	slog.Debug("Updated power gauge", "chargeRate", info.ChargeRate, "text", powerText)
}

// updateHealthGauge updates the health gauge display
func (v *View) updateHealthGauge(info *battery.Info) {
	healthPercent := info.Health()
	healthColor := getHealthColor(healthPercent)
	healthBar := CreateProgressBar(healthPercent, ProgressBarWidth, ProgressBarStyleASCII)
	healthText := fmt.Sprintf(" [%s]%s[-] [%s]%.1f%%[-]", healthColor, healthBar, healthColor, healthPercent)
	v.healthGauge.SetText(healthText)
	slog.Debug("Updated health gauge", "percent", healthPercent, "text", healthText)
}

// updateCharts updates the chart display
func (v *View) updateCharts() {
	if !v.validateChartDimensions() {
		return
	}

	var fullText strings.Builder
	v.renderChartTitle(&fullText)
	v.renderChartContent(&fullText)

	v.chartArea.Clear()
	v.chartArea.SetText(fullText.String())
}

// validateChartDimensions checks if chart dimensions are valid
func (v *View) validateChartDimensions() bool {
	if v.chartWidth <= 0 || v.chartHeight <= 0 {
		slog.Debug("Skipping chart update - invalid dimensions",
			"width", v.chartWidth,
			"height", v.chartHeight)
		return false
	}

	slog.Debug("Updating charts", "width", v.chartWidth, "height", v.chartHeight)
	return true
}

// renderChartTitle renders the chart title with decorative borders
func (v *View) renderChartTitle(text *strings.Builder) {
	const title = " Real-time Monitoring "
	titleLen := len(title)

	if v.chartWidth <= titleLen {
		return
	}

	leftPadding := (v.chartWidth - titleLen) / 2
	rightPadding := v.chartWidth - leftPadding - titleLen

	titleLine := fmt.Sprintf("[white::b]%s%s%s[-]",
		strings.Repeat("─", leftPadding),
		title,
		strings.Repeat("─", rightPadding))

	text.WriteString(titleLine)
	text.WriteString("\n")
}

// renderChartContent renders the actual chart data
func (v *View) renderChartContent(text *strings.Builder) {
	// Update chart sizes (account for title)
	v.chartSet.SetSize(v.chartWidth, v.chartHeight-1)

	// Render charts
	chartText := v.chartSet.Render()
	if chartText == "" {
		slog.Warn("Chart render returned empty string")
		return
	}

	// Debug logging
	lines := strings.Split(chartText, "\n")
	slog.Debug("Chart rendered", "lines", len(lines))
	if len(lines) > 0 {
		slog.Debug("First line", "content", lines[0])
	}
	if len(lines) > 1 {
		slog.Debug("Second line", "content", lines[1])
	}

	text.WriteString(chartText)
}

// Helper functions

func getChargeColor(percent float64) string {
	return GetColorByThreshold(percent, ColorThresholdsDefault)
}

func getHealthColor(percent float64) string {
	return GetColorByThreshold(percent, ColorThresholdsHealth)
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
