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
	Field string
	Value interface{}
	Err   error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config field %s (value: %v): %v", e.Field, e.Value, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new configuration error
func NewConfigError(field string, value interface{}, err error) error {
	return &ConfigError{
		Field: field,
		Value: value,
		Err:   err,
	}
}
