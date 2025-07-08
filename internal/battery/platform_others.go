//go:build !linux

package battery

import pkgErrors "github.com/xsikor/go-battop/internal/errors"

type defaultPlatformReader struct{}

func newPlatformReader() PlatformReader {
	return &defaultPlatformReader{}
}

// ReadBatteryStats returns empty stats on non-Linux platforms
func (r *defaultPlatformReader) ReadBatteryStats(batteryIndex int) (BatteryStats, error) {
	// Return error indicating platform is not supported
	return BatteryStats{}, pkgErrors.ErrPlatformNotSupported
}
