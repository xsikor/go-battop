package app

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/rivo/tview"
	"github.com/xsikor/go-battop/internal/battery"
	"github.com/xsikor/go-battop/internal/ui"
)

// Application orchestrates the battery monitoring terminal UI application
type Application struct {
	config   *Config
	tviewApp *tview.Application
	manager  *battery.Manager
	events   *EventManager
	ui       interface {
		GetRoot() tview.Primitive
		Update() error
		NextTab()
		PreviousTab()
	}
}

// New creates and initializes a new Application with the given configuration
func New(config *Config) *Application {
	return &Application{
		config:   config,
		tviewApp: tview.NewApplication(),
		manager:  battery.NewManager(),
	}
}

// Run starts the main application event loop and blocks until exit
func (a *Application) Run() error {
	slog.Info("Starting battop", "version", "0.3.0")

	// Initial battery update
	if err := a.manager.Update(); err != nil {
		return fmt.Errorf("initial battery update failed: %w", err)
	}

	// Check if we have batteries
	batteries, err := a.manager.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get batteries: %w", err)
	}
	if len(batteries) == 0 {
		return fmt.Errorf("no batteries found on this system")
	}

	slog.Info("Found batteries", "count", len(batteries))

	// Create UI
	ui, err := ui.NewInterface(a.manager, a.config)
	if err != nil {
		return fmt.Errorf("failed to create UI: %w", err)
	}
	a.ui = ui

	// Set up event manager
	a.events = NewEventManager(a.tviewApp, a.config)
	a.events.Start()
	defer a.events.Stop()

	// Set root and enable mouse
	root := a.ui.GetRoot()
	if root == nil {
		return fmt.Errorf("UI root is nil")
	}

	slog.Info("Setting up tview application")
	a.tviewApp.SetRoot(root, true).SetFocus(root)
	a.tviewApp.EnableMouse(true)

	// Check terminal size after initial setup
	go func() {
		time.Sleep(100 * time.Millisecond)
		a.tviewApp.QueueUpdateDraw(func() {
			// Terminal size will be logged during resize events
			slog.Info("Initial UI setup complete")
		})
	}()

	// Start event processing in separate goroutine
	go a.processEvents()

	// Force initial UI update and draw
	if err := a.ui.Update(); err != nil {
		slog.Warn("Initial UI update failed", "error", err)
	}

	// Schedule an initial draw after a short delay to ensure proper rendering
	go func() {
		time.Sleep(50 * time.Millisecond)
		a.tviewApp.QueueUpdateDraw(func() {})
	}()

	slog.Info("Starting tview main loop")

	// Run the tview application (blocks)
	if err := a.tviewApp.Run(); err != nil {
		return fmt.Errorf("tview error: %w", err)
	}

	return nil
}

// processEvents processes application events
func (a *Application) processEvents() {
	for event := range a.events.Events() {
		switch event.Type {
		case EventExit:
			slog.Info("Exit event received")
			a.tviewApp.Stop()
			return

		case EventNextTab:
			slog.Debug("Next tab event")
			a.ui.NextTab()
			a.tviewApp.Draw()

		case EventPreviousTab:
			slog.Debug("Previous tab event")
			a.ui.PreviousTab()
			a.tviewApp.Draw()

		case EventTick:
			// Update battery information
			if err := a.manager.Update(); err != nil {
				slog.Error("Failed to update batteries",
					"error", err,
					"battery_count", a.manager.Count(),
					"update_interval", a.config.Delay,
				)
				// Don't exit on update errors, just log them
			}

			// Update UI
			if err := a.ui.Update(); err != nil {
				slog.Error("Failed to update UI",
					"error", err,
					"battery_count", a.manager.Count(),
				)
			}

			// Redraw
			a.tviewApp.Draw()

		case EventResize:
			slog.Debug("Resize event")
			a.tviewApp.Draw()
		}
	}
}
