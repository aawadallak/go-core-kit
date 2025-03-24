package logger

import (
	"fmt"
	"log"
)

// consoleWriter implements the Provider interface for console-based logging
type consoleWriter struct {
	minLevel Level
}

// Ensure Provider interface implementation
var _ Provider = (*consoleWriter)(nil)

// initConsoleLogger creates a new console logging provider
func initConsoleLogger() Provider {
	return &consoleWriter{minLevel: DefaultLevel()}
}

// Output writes a log message with level and optional attributes
func (cw *consoleWriter) Write(rank Level, content string, modifiers ...Option) {
	config := NewConfig(modifiers...)

	formatStr := "[%s] text=%s"
	values := []interface{}{
		rank.String(),
		content,
	}

	for attr, val := range config.Attributes {
		switch v := val.(type) {
		case string:
			formatStr += fmt.Sprintf(" %s=%s", attr, v)
		case int, int8, int16, int32, int64:
			formatStr += fmt.Sprintf(" %s=%d", attr, v)
		case uint, uint8, uint16, uint32, uint64:
			formatStr += fmt.Sprintf(" %s=%d", attr, v)
		// TODO: Consider handling complex nested structures
		default:
			// Ignore unsupported types
		}
	}

	log.Printf(formatStr, values...)
}

// OutputFormatted writes a formatted message with level prefix
func (cw *consoleWriter) WriteF(rank Level, layout string, params ...interface{}) {
	entry := fmt.Sprintf("[%s] - %s", rank.String(), layout)
	if len(params) > 0 {
		log.Printf(entry, params...)
		return
	}

	log.Print(entry)
}

// IsActive checks if the specified level is enabled
func (cw *consoleWriter) Enabled(rank Level) bool {
	return rank >= cw.minLevel
}
