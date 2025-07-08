//go:build linux

package battery

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type linuxPlatformReader struct{}

func newPlatformReader() PlatformReader {
	return &linuxPlatformReader{}
}

// ReadBatteryStats reads battery statistics from Linux sysfs
func (r *linuxPlatformReader) ReadBatteryStats(batteryIndex int) (BatteryStats, error) {
	stats := BatteryStats{}

	// Find battery path
	batteryPath := fmt.Sprintf("/sys/class/power_supply/BAT%d", batteryIndex)

	// Check if battery exists
	if _, err := os.Stat(batteryPath); os.IsNotExist(err) {
		// Try alternative naming
		batteryPath = fmt.Sprintf("/sys/class/power_supply/BAT%d", batteryIndex)
		if _, err := os.Stat(batteryPath); os.IsNotExist(err) {
			return stats, fmt.Errorf("battery %d not found", batteryIndex)
		}
	}

	// Read cycle count
	if cycleCount, err := readSysfsInt(filepath.Join(batteryPath, "cycle_count")); err == nil {
		stats.CycleCount = cycleCount
	}

	// Read manufacturer
	if manufacturer, err := readSysfsString(filepath.Join(batteryPath, "manufacturer")); err == nil {
		stats.Manufacturer = manufacturer
	}

	// Read model name
	if modelName, err := readSysfsString(filepath.Join(batteryPath, "model_name")); err == nil {
		stats.ModelName = modelName
	}

	// Read serial number
	if serial, err := readSysfsString(filepath.Join(batteryPath, "serial_number")); err == nil {
		stats.SerialNumber = serial
	}

	// Read technology
	if technology, err := readSysfsString(filepath.Join(batteryPath, "technology")); err == nil {
		stats.Technology = technology
	}

	return stats, nil
}

// readSysfsString reads a string value from a sysfs file
func readSysfsString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// readSysfsInt reads an integer value from a sysfs file
func readSysfsInt(path string) (int, error) {
	str, err := readSysfsString(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}
