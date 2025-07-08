package app

import (
	"flag"
	"fmt"
	"time"

	"github.com/xsikor/go-battop/internal/errors"
)

// Units defines the measurement unit system for displaying battery values
type Units string

const (
	// UnitsHuman displays values in human-readable units (W, Wh)
	UnitsHuman Units = "human"
	// UnitsRaw displays values in raw units (mW, mWh)
	UnitsRaw Units = "raw"
)

// Config defines the application configuration parameters
type Config struct {
	// Delay between updates
	Delay time.Duration

	// Units to use for display
	Units Units

	// Verbose enables debug logging
	Verbose bool

	// Version flag
	Version bool
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Delay:   1 * time.Second,
		Units:   UnitsHuman,
		Verbose: false,
		Version: false,
	}
}

// ParseFlags parses command line flags and returns configuration
func ParseFlags() (*Config, error) {
	config := DefaultConfig()

	var delayStr string
	var unitsStr string

	flag.StringVar(&delayStr, "delay", "1s", "Delay between updates (e.g., 1s, 500ms)")
	flag.StringVar(&unitsStr, "units", "human", "Units to use (human: W/Wh, raw: mW/mWh)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&config.Version, "version", false, "Show version and exit")

	flag.Parse()

	// Parse delay
	if delayStr != "" {
		delay, err := time.ParseDuration(delayStr)
		if err != nil {
			return nil, errors.NewConfigError("delay", delayStr, err)
		}
		if delay < 100*time.Millisecond {
			return nil, errors.NewConfigError("delay", delay, fmt.Errorf("delay must be at least 100ms"))
		}
		config.Delay = delay
	}

	// Parse units
	switch unitsStr {
	case "human", "h":
		config.Units = UnitsHuman
	case "raw", "r":
		config.Units = UnitsRaw
	default:
		return nil, errors.NewConfigError("units", unitsStr, fmt.Errorf("invalid units: must be 'human' or 'raw'"))
	}

	return config, nil
}

// FormatPower formats power value according to units setting
func (c *Config) FormatPower(mW float64) string {
	if c.Units == UnitsHuman {
		return fmt.Sprintf("%.2f W", mW/1000.0)
	}
	return fmt.Sprintf("%.0f mW", mW)
}

// FormatEnergy formats energy value according to units setting
func (c *Config) FormatEnergy(mWh float64) string {
	if c.Units == UnitsHuman {
		return fmt.Sprintf("%.2f Wh", mWh/1000.0)
	}
	return fmt.Sprintf("%.0f mWh", mWh)
}

// FormatVoltage formats voltage value
func (c *Config) FormatVoltage(v float64) string {
	return fmt.Sprintf("%.2f V", v)
}
