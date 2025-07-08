package battery

import (
	"time"
)

// State represents battery state
type State int

const (
	StateUnknown State = iota
	StateEmpty
	StateFull
	StateCharging
	StateDischarging
	StateNotCharging
)

// String returns string representation of battery state
func (s State) String() string {
	switch s {
	case StateEmpty:
		return "Empty"
	case StateFull:
		return "Full"
	case StateCharging:
		return "Charging"
	case StateDischarging:
		return "Discharging"
	case StateNotCharging:
		return "Not charging"
	default:
		return "Unknown"
	}
}

// Info contains battery information
type Info struct {
	// Index is the battery index (0-based)
	Index int

	// State is the current battery state
	State State

	// Current capacity in mWh
	Current float64

	// Full capacity in mWh (last full charge)
	Full float64

	// Design capacity in mWh
	Design float64

	// Charge rate in mW (positive = charging, negative = discharging)
	ChargeRate float64

	// Voltage in V
	Voltage float64

	// Design voltage in V
	DesignVoltage float64

	// Cycle count (if available)
	CycleCount int

	// Technology (e.g., "Li-ion")
	Technology string

	// Serial number
	Serial string

	// Model name
	Model string

	// Manufacturer
	Manufacturer string

	// Temperature in Celsius (if available)
	Temperature float64

	// Last update time
	UpdatedAt time.Time
}

// ChargePercent returns the current charge percentage
func (b *Info) ChargePercent() float64 {
	if b.Full <= 0 {
		return 0
	}
	percent := (b.Current / b.Full) * 100
	if percent > 100 {
		return 100
	}
	if percent < 0 {
		return 0
	}
	return percent
}

// Health returns battery health percentage (full capacity vs design capacity)
func (b *Info) Health() float64 {
	if b.Design <= 0 {
		return 0
	}
	health := (b.Full / b.Design) * 100
	if health > 100 {
		return 100
	}
	if health < 0 {
		return 0
	}
	return health
}

// TimeToEmpty estimates time until battery is empty (during discharge)
func (b *Info) TimeToEmpty() time.Duration {
	if b.ChargeRate >= 0 || b.Current <= 0 {
		return 0
	}
	hours := b.Current / (-b.ChargeRate)
	return time.Duration(hours * float64(time.Hour))
}

// TimeToFull estimates time until battery is full (during charge)
func (b *Info) TimeToFull() time.Duration {
	if b.ChargeRate <= 0 || b.Full <= b.Current {
		return 0
	}
	hours := (b.Full - b.Current) / b.ChargeRate
	return time.Duration(hours * float64(time.Hour))
}
