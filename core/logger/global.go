package logger

import (
	"sync"
)

// loggerInstance manages the global Logger instance with thread-safe access
type loggerInstance struct {
	logger Logger     // The global logger instance
	mu     sync.Mutex // Protects logger during writes
	once   sync.Once  // Ensures single initialization
}

// global holds the singleton logger instance
var global = &loggerInstance{}

// Compile-time check for Logger interface implementation
var _ Logger = (*logger)(nil)

// Global returns the global logger instance
// If not yet initialized, it creates a default instance
// This function is thread-safe and always returns a non-nil Logger
func Global() Logger {
	if global.logger == nil {
		global.once.Do(func() {
			global.logger = New()
		})
	}
	return global.logger
}

// SetInstance sets the global logger instance to the provided logger
// If the provided logger is nil, it initializes with a default instance
// This function is thread-safe and can be called multiple times
func SetInstance(logger Logger) {
	if logger == nil {
		// Ensure we have at least a default logger
		global.once.Do(func() {
			global.logger = New()
		})
		return
	}

	global.mu.Lock()
	defer global.mu.Unlock()
	global.logger = logger
}
