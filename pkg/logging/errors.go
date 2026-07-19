package logging

import "errors"

var (
	// ErrUnknownLogLevel is returned when the log level string is not recognized.
	ErrUnknownLogLevel = errors.New("unknown log level (valid: debug, info, warn, error)")
	// ErrUnknownLogFormat is returned when the log format string is not recognized.
	ErrUnknownLogFormat = errors.New("unknown log format (valid: console, text, json)")
)
