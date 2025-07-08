package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/xsikor/go-battop/internal/app"
)

var (
	version = "0.3.0"
	commit  = "development"
	date    = "unknown"
)

func main() {
	// Parse configuration
	config, err := app.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle version flag
	if config.Version {
		fmt.Printf("battop %s (%s) built on %s\n", version, commit, date)
		os.Exit(0)
	}

	// Set up logging
	logLevel := slog.LevelInfo
	if config.Verbose {
		logLevel = slog.LevelDebug
	}

	// Create or open error log file in temp directory
	logPath := filepath.Join(os.TempDir(), "go-battop.log")
	errorLog, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open error log at %s: %v\n", logPath, err)
		os.Exit(1)
	}
	defer errorLog.Close()

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	// Log to error.log file
	logger := slog.New(slog.NewTextHandler(errorLog, opts))
	slog.SetDefault(logger)

	// Create and run application
	application := app.New(config)
	if err := application.Run(); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
}
