//go:build !linux

package battery

type defaultPlatformReader struct{}

func newPlatformReader() PlatformReader {
	return &defaultPlatformReader{}
}

// ReadBatteryStats returns empty stats on non-Linux platforms
func (r *defaultPlatformReader) ReadBatteryStats(batteryIndex int) (BatteryStats, error) {
	// Return empty stats for unsupported platforms
	return BatteryStats{}, nil
}
