package battery

// PlatformReader reads platform-specific battery information
type PlatformReader interface {
	// ReadBatteryStats reads additional battery statistics not provided by distatus/battery
	// Returns cycle count and any errors encountered
	ReadBatteryStats(batteryIndex int) (stats BatteryStats, err error)
}

// BatteryStats contains platform-specific battery statistics
type BatteryStats struct {
	// CycleCount is the number of charge cycles the battery has gone through
	CycleCount int

	// Manufacturer of the battery
	Manufacturer string

	// ModelName of the battery
	ModelName string

	// SerialNumber of the battery
	SerialNumber string

	// Technology type (e.g., "Li-ion", "Li-poly")
	Technology string
}

// GetPlatformReader returns a platform-specific battery reader
func GetPlatformReader() PlatformReader {
	return newPlatformReader()
}
