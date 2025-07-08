package battery

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/distatus/battery"
	"github.com/xsikor/go-battop/internal/errors"
)

// Manager manages battery information
type Manager struct {
	mu             sync.RWMutex
	batteries      []*Info
	lastError      error
	platformReader PlatformReader
}

// NewManager creates a new battery manager
func NewManager() *Manager {
	return &Manager{
		batteries:      make([]*Info, 0),
		platformReader: GetPlatformReader(),
	}
}

// Update updates battery information
func (m *Manager) Update() error {
	batteries, err := battery.GetAll()
	if err != nil {
		m.mu.Lock()
		m.lastError = err
		m.mu.Unlock()
		return fmt.Errorf("failed to get batteries: %w", err)
	}

	if len(batteries) == 0 {
		m.mu.Lock()
		m.lastError = errors.ErrNoBatteries
		m.mu.Unlock()
		return errors.ErrNoBatteries
	}

	infos := make([]*Info, 0, len(batteries))
	now := time.Now()

	for i, bat := range batteries {
		info := &Info{
			Index:         i,
			State:         convertState(bat.State),
			Current:       bat.Current,
			Full:          bat.Full,
			Design:        bat.Design,
			ChargeRate:    bat.ChargeRate,
			Voltage:       bat.Voltage,
			DesignVoltage: bat.DesignVoltage,
			UpdatedAt:     now,
		}

		// Read platform-specific battery stats
		if platformStats, err := m.platformReader.ReadBatteryStats(i); err == nil {
			info.CycleCount = platformStats.CycleCount
			if platformStats.Technology != "" {
				info.Technology = platformStats.Technology
			} else {
				info.Technology = "Li-ion" // Default if not available
			}
			if platformStats.Manufacturer != "" {
				info.Manufacturer = platformStats.Manufacturer
			}
			if platformStats.ModelName != "" {
				info.Model = platformStats.ModelName
			}
			if platformStats.SerialNumber != "" {
				info.Serial = platformStats.SerialNumber
			}
		} else {
			// Set defaults if platform stats not available
			info.Technology = "Li-ion"
			slog.Debug("Failed to read platform battery stats",
				"index", i,
				"error", err,
			)
		}

		// Temperature is not directly available in distatus/battery
		info.Temperature = 0

		// For charge rate, ensure proper sign based on state
		if info.State == StateDischarging && info.ChargeRate > 0 {
			info.ChargeRate = -info.ChargeRate
		}

		infos = append(infos, info)

		slog.Debug("Updated battery info",
			"index", i,
			"state", info.State.String(),
			"current", info.Current,
			"full", info.Full,
			"charge_rate", info.ChargeRate,
			"voltage", info.Voltage,
		)
	}

	m.mu.Lock()
	m.batteries = infos
	m.lastError = nil
	m.mu.Unlock()

	return nil
}

// GetAll returns all battery information
func (m *Manager) GetAll() ([]*Info, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.lastError != nil {
		return nil, m.lastError
	}

	// Return a copy to prevent data races
	result := make([]*Info, len(m.batteries))
	for i, bat := range m.batteries {
		// Create a copy of the battery info
		batCopy := *bat
		result[i] = &batCopy
	}

	return result, nil
}

// Get returns battery information by index
func (m *Manager) Get(index int) (*Info, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.lastError != nil {
		return nil, m.lastError
	}

	if index < 0 || index >= len(m.batteries) {
		return nil, errors.ErrBatteryNotFound
	}

	// Return a copy to prevent data races
	batCopy := *m.batteries[index]
	return &batCopy, nil
}

// Count returns the number of batteries
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.batteries)
}

// convertState converts distatus/battery state to our state
func convertState(s battery.State) State {
	switch s.String() {
	case "Empty":
		return StateEmpty
	case "Full":
		return StateFull
	case "Charging":
		return StateCharging
	case "Discharging":
		return StateDischarging
	case "Not charging":
		return StateNotCharging
	default:
		return StateUnknown
	}
}
