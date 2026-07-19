// Package logging provides configured slog loggers with colorful console
// output for development and structured JSON for production.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/lmittmann/tint"
)

// Format is the log output format.
type Format string

const (
	// FormatConsole produces colorful, human-readable output for development.
	FormatConsole Format = "console"
	// FormatText is an alias for FormatConsole.
	FormatText Format = "text"
	// FormatJSON produces structured JSON output for production.
	FormatJSON Format = "json"
)

// ParseLevel parses a level string (debug, info, warn, error) into a slog.Level.
func ParseLevel(levelStr string) (slog.Level, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("%w: %s", ErrUnknownLogLevel, levelStr)
	}
}

// ParseFormat parses a format string (console, text, json) into a Format.
// console and text are equivalent: both produce colorful, human-readable output.
func ParseFormat(formatStr string) (Format, error) {
	switch strings.ToLower(formatStr) {
	case "console":
		return FormatConsole, nil
	case "text":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnknownLogFormat, formatStr)
	}
}

// New creates a new slog.Logger with the given level and format, writing to
// writer. For console/text format, tint automatically disables colors when
// writer is not a terminal.
func New(level slog.Level, format Format, writer io.Writer) *slog.Logger {
	var handler slog.Handler

	switch format {
	case FormatJSON:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	case FormatConsole, FormatText:
		handler = tint.NewTextHandler(writer, &tint.Options{Level: level})
	default:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}
