package ui

import (
	"fmt"
	"log/slog"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xsikor/go-battop/internal/battery"
	"github.com/xsikor/go-battop/internal/errors"
)

// Config interface to avoid circular imports
type Config interface {
	FormatPower(mW float64) string
	FormatEnergy(mWh float64) string
	FormatVoltage(v float64) string
}

// Interface manages the main UI
type Interface struct {
	root    *tview.Flex
	view    *View
	manager *battery.Manager
	config  Config
}

// NewInterface creates a new UI interface
func NewInterface(manager *battery.Manager, config Config) (*Interface, error) {
	if manager == nil {
		return nil, fmt.Errorf("battery manager is nil")
	}

	i := &Interface{
		manager: manager,
		config:  config,
	}

	// Initialize first battery only
	if err := i.initializeBattery(); err != nil {
		return nil, err
	}

	// Build UI layout
	i.buildLayout()

	return i, nil
}

// GetRoot returns the root UI element
func (i *Interface) GetRoot() tview.Primitive {
	return i.root
}

// initializeBattery initializes the first battery view
func (i *Interface) initializeBattery() error {
	batteries, err := i.manager.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get batteries: %w", err)
	}

	if len(batteries) == 0 {
		return errors.ErrNoBatteries
	}

	// Create a view for the first battery only
	bat := batteries[0]
	i.view = NewView(bat.Index, i.config)
	i.view.Update(bat)

	slog.Info("Initialized battery view", "index", bat.Index)
	return nil
}

// buildLayout builds the UI layout
func (i *Interface) buildLayout() {
	// Create main container
	container := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add the battery view - takes all space except footer
	container.AddItem(i.view.GetRoot(), 0, 1, true)

	// Add help footer
	helpText := tview.NewTextView()
	helpText.SetDynamicColors(true)
	helpText.SetTextAlign(tview.AlignCenter)
	helpText.SetBackgroundColor(tcell.ColorDefault)
	helpText.SetText("[gray]Press [yellow]q[gray]/[yellow]ESC[gray] to quit[-]")
	container.AddItem(helpText, 1, 0, false)

	i.root = container
}

// Update updates the UI with latest battery information
func (i *Interface) Update() error {
	batteries, err := i.manager.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get batteries: %w", err)
	}

	// Update the first battery view
	if len(batteries) > 0 {
		i.view.Update(batteries[0])
	}

	return nil
}

// NextTab is no longer needed but kept for interface compatibility
func (i *Interface) NextTab() {
	// No-op
}

// PreviousTab is no longer needed but kept for interface compatibility
func (i *Interface) PreviousTab() {
	// No-op
}
