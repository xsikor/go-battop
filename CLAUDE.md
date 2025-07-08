# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

**battop** is a terminal-based battery monitoring tool (similar to `top` or `htop`) that has been rewritten from Rust to Go. The application provides real-time battery statistics with an interactive terminal UI featuring progress bars, time-series charts, and multi-battery support.

## Build and Development Commands

### Essential Commands
```bash
# Build the application
make build

# Run the application
make run

# Run with verbose logging
make run-verbose

# Format code (required before commits)
make fmt

# Run linter (must pass with zero issues)
make lint

# Run tests (though no tests implemented yet)
make test

# Build optimized release version
make release

# Development mode with auto-reload (requires air)
make dev
```

### Running with Options
```bash
# Custom update interval
./build/battop -delay 2s

# Raw units (mW/mWh instead of W/Wh)
./build/battop -units raw

# Enable verbose logging
./build/battop -verbose

# Show version
./build/battop -version
```

## Architecture and Code Structure

### Event-Driven Architecture
The application uses a channel-based event system to coordinate between components:
- **Event Types**: Exit, NextTab, PreviousTab, Tick (defined in `internal/app/events.go`)
- **Event Manager**: Routes events between UI, battery updates, and application logic
- **Goroutines**: Separate goroutines handle UI rendering, battery polling, and input events

### Component Interaction Flow
```
main.go → Application → EventManager → UI Interface
                    ↓                      ↓
              BatteryManager          Battery Views
                    ↓                      ↓
           distatus/battery            Charts/Widgets
```

### Key Design Patterns

1. **Interface-Based Design**: To avoid circular dependencies between packages
   - `ui.Interface` defines UI contract
   - `app.EventHandler` for event handling
   - `battery.Manager` abstracts battery operations

2. **Thread-Safe Data Access**: Battery data protected by mutex
   - All battery state updates go through `battery.Manager`
   - UI reads data through safe getter methods

3. **Separation of Concerns**:
   - `cmd/`: Entry point and CLI parsing only
   - `internal/app/`: Application orchestration and business logic
   - `internal/ui/`: All terminal UI components
   - `internal/battery/`: Battery data management
   - `internal/errors/`: Custom error types

### UI Component Architecture

The UI is built with `rivo/tview` and consists of:
- **Interface** (`ui/interface.go`): Main container managing tabs and layout
- **View** (`ui/view.go`): Individual battery display with:
  - Left panel: Battery info, progress bars for charge/health
  - Right panel: Time-series charts
- **Charts** (`ui/charts.go`): Custom ASCII chart rendering engine
  - Maintains history buffers for each metric
  - Auto-scales Y-axis based on data range
  - Handles time-based X-axis with markers

### Critical Implementation Details

1. **Battery Updates**: 
   - Polling runs in separate goroutine
   - Updates trigger UI refresh via event system
   - History maintained for chart visualization

2. **Chart Rendering**:
   - Custom ASCII implementation due to tview limitations
   - Uses braille characters for smooth lines
   - Auto-scaling with proper axis labels

3. **Error Handling**:
   - Custom error types in `internal/errors/`
   - Errors logged to `error.log` with slog
   - UI continues running on non-fatal errors

4. **Keyboard Input**:
   - Handled by tview's event system
   - Mapped to internal events via `EventManager`
   - Supports vim-style navigation (h/l for tabs)

## Important Notes

- **No Tests Yet**: The codebase has no test files implemented
- **Cross-Platform**: Uses `distatus/battery` for OS abstraction
- **Logging**: All errors logged to `error.log` file
- **Configuration**: Currently via CLI flags only (no config file)
- **Dependencies**: Minimal - only tview and distatus/battery

## Common Development Tasks

When adding new features:
1. Update event types if new user interactions needed
2. Extend `battery.Manager` for new battery data
3. Modify `ui.View` for new display elements
4. Update chart rendering for new metrics
5. Always run `make fmt && make lint` before committing