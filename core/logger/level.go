package logger

import (
	"os"
	"strings"
)

// Level represents logging severity levels
type Level int8

const (
	DebugLevel Level = iota // 0
	InfoLevel               // 1
	WarnLevel               // 2
	ErrorLevel              // 3
	FatalLevel              // 4
)

var (
	defaultLevel   = DebugLevel
	levelStringMap = map[string]Level{
		"DEBUG": DebugLevel,
		"INFO":  InfoLevel,
		"WARN":  WarnLevel,
		"ERROR": ErrorLevel,
		"FATAL": FatalLevel,
	}
)

// String returns the string representation of a Level
func (l Level) String() string {
	for s, lvl := range levelStringMap {
		if lvl == l {
			return s
		}
	}
	return "UNKNOWN"
}

// DefaultLevel returns the configured level from LOGGER_LEVEL environment variable
func DefaultLevel() Level {
	return fromString(os.Getenv("LOGGER_LEVEL"))
}

// fromString converts a string to a Level, returning defaultLevel if invalid
func fromString(level string) Level {
	if level == "" {
		return defaultLevel
	}
	if l, ok := levelStringMap[strings.ToUpper(level)]; ok {
		return l
	}
	return defaultLevel
}

// IsGreaterThan compares two levels
func (l Level) IsGreaterThan(other Level) bool {
	return l > other
}
