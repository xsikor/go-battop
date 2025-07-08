package app

import (
	"log/slog"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// EventType represents the type of event
type EventType int

const (
	// EventExit signals application exit
	EventExit EventType = iota

	// EventNextTab switches to next battery tab
	EventNextTab

	// EventPreviousTab switches to previous battery tab
	EventPreviousTab

	// EventTick signals a periodic update
	EventTick

	// EventResize signals terminal resize
	EventResize
)

// Event represents an application event
type Event struct {
	Type EventType
}

// EventManager manages application events
type EventManager struct {
	app       *tview.Application
	eventChan chan Event
	stopChan  chan struct{}
	config    *Config
}

// NewEventManager creates a new event manager
func NewEventManager(app *tview.Application, config *Config) *EventManager {
	return &EventManager{
		app:       app,
		eventChan: make(chan Event, EventChannelBufferSize),
		stopChan:  make(chan struct{}),
		config:    config,
	}
}

// Start starts the event manager
func (em *EventManager) Start() {
	// Start tick timer
	go em.tickLoop()

	// Set up keyboard handlers
	em.setupKeyboardHandlers()
}

// Stop stops the event manager
func (em *EventManager) Stop() {
	close(em.stopChan)
}

// Events returns the event channel
func (em *EventManager) Events() <-chan Event {
	return em.eventChan
}

// tickLoop generates periodic tick events
func (em *EventManager) tickLoop() {
	ticker := time.NewTicker(em.config.Delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case em.eventChan <- Event{Type: EventTick}:
				slog.Debug("Tick event sent")
			default:
				slog.Warn("Event channel full, dropping tick event")
			}
		case <-em.stopChan:
			return
		}
	}
}

// setupKeyboardHandlers sets up keyboard event handlers
func (em *EventManager) setupKeyboardHandlers() {
	em.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			em.sendEvent(Event{Type: EventExit})
			return nil
		case tcell.KeyTab, tcell.KeyRight:
			em.sendEvent(Event{Type: EventNextTab})
			return nil
		case tcell.KeyBacktab, tcell.KeyLeft:
			em.sendEvent(Event{Type: EventPreviousTab})
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				em.sendEvent(Event{Type: EventExit})
				return nil
			case 'h', 'H':
				em.sendEvent(Event{Type: EventPreviousTab})
				return nil
			case 'l', 'L':
				em.sendEvent(Event{Type: EventNextTab})
				return nil
			}
		}
		return event
	})
}

// sendEvent sends an event to the event channel
func (em *EventManager) sendEvent(event Event) {
	select {
	case em.eventChan <- event:
		slog.Debug("Event sent", "type", event.Type)
	default:
		slog.Warn("Event channel full, dropping event", "type", event.Type)
	}
}
