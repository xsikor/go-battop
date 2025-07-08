package errors

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrNoBatteries is returned when no batteries are found on the system
	ErrNoBatteries = errors.New("no batteries found")

	// ErrBatteryNotFound is returned when a specific battery is not found
	ErrBatteryNotFound = errors.New("battery not found")

	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrUIInit is returned when UI initialization fails
	ErrUIInit = errors.New("failed to initialize UI")

	// ErrPlatformNotSupported is returned when platform-specific features are not available
	ErrPlatformNotSupported = errors.New("platform not supported")

	// ErrFeatureNotAvailable is returned when a feature is not available on the current platform
	ErrFeatureNotAvailable = errors.New("feature not available on this platform")
)

// BatteryError represents a battery-specific error
type BatteryError struct {
	Index int
	Op    string
	Err   error
}

func (e *BatteryError) Error() string {
	return fmt.Sprintf("battery %d: %s: %v", e.Index, e.Op, e.Err)
}

func (e *BatteryError) Unwrap() error {
	return e.Err
}

// NewBatteryError creates a new battery error
func NewBatteryError(index int, op string, err error) error {
	return &BatteryError{
		Index: index,
		Op:    op,
		Err:   err,
	}
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field    string
	ValueStr string
	Err      error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config field %s (value: %s): %v", e.Field, e.ValueStr, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new configuration error
func NewConfigError(field string, value interface{}, err error) error {
	return &ConfigError{
		Field:    field,
		ValueStr: fmt.Sprintf("%v", value),
		Err:      err,
	}
}
