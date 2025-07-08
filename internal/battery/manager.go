package battery

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/distatus/battery"
	pkgErrors "github.com/xsikor/go-battop/internal/errors"
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
	// ATTN: Early validation reduces nesting and improves readability
	batteries, err := battery.GetAll()
	if err != nil {
		return m.setLastError(fmt.Errorf("failed to get batteries: %w", err))
	}

	if len(batteries) == 0 {
		return m.setLastError(pkgErrors.ErrNoBatteries)
	}

	// Happy path: convert and update battery information
	infos := m.convertBatteriesToInfo(batteries)

	m.mu.Lock()
	m.batteries = infos
	m.lastError = nil
	m.mu.Unlock()

	return nil
}

// convertBatteriesToInfo converts battery.Battery objects to our Info structs
func (m *Manager) convertBatteriesToInfo(batteries []*battery.Battery) []*Info {
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
			Temperature:   0, // Not directly available in distatus/battery
		}

		// Enrich with platform-specific data
		m.enrichBatteryWithPlatformStats(info, i)

		// Ensure charge rate sign is correct
		m.normalizeChargeRate(info)

		infos = append(infos, info)

		// Log the update
		m.logBatteryUpdate(info, i)
	}

	return infos
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
		return nil, pkgErrors.ErrBatteryNotFound
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

// setLastError sets the last error with proper locking
func (m *Manager) setLastError(err error) error {
	m.mu.Lock()
	m.lastError = err
	m.mu.Unlock()
	return err
}

// normalizeChargeRate ensures charge rate sign matches battery state
func (m *Manager) normalizeChargeRate(info *Info) {
	if info.State == StateDischarging && info.ChargeRate > 0 {
		info.ChargeRate = -info.ChargeRate
	}
}

// logBatteryUpdate logs battery update information
func (m *Manager) logBatteryUpdate(info *Info, index int) {
	slog.Debug("Updated battery info",
		"index", index,
		"state", info.State.String(),
		"current", info.Current,
		"full", info.Full,
		"charge_rate", info.ChargeRate,
		"voltage", info.Voltage,
	)
}

// enrichBatteryWithPlatformStats applies platform-specific stats to battery info
func (m *Manager) enrichBatteryWithPlatformStats(info *Info, index int) {
	platformStats, err := m.platformReader.ReadBatteryStats(index)
	if err != nil {
		// Set defaults if platform stats not available
		info.Technology = "Li-ion"

		// Log appropriately based on error type
		if errors.Is(err, pkgErrors.ErrPlatformNotSupported) {
			slog.Debug("Platform-specific stats not available",
				"index", index,
				"platform", "non-linux",
			)
		} else {
			slog.Warn("Failed to read platform battery stats",
				"index", index,
				"error", err,
			)
		}
		return
	}

	// Apply available stats
	info.CycleCount = platformStats.CycleCount

	// Set technology with default fallback
	info.Technology = coalesce(platformStats.Technology, "Li-ion")

	// Set other fields if available
	if platformStats.Manufacturer != "" {
		info.Manufacturer = platformStats.Manufacturer
	}
	if platformStats.ModelName != "" {
		info.Model = platformStats.ModelName
	}
	if platformStats.SerialNumber != "" {
		info.Serial = platformStats.SerialNumber
	}
}

// coalesce returns the first non-empty string
func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
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
