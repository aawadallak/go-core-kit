package logger

// Logger defines the interface for logging operations at different severity levels.
// Each level provides three methods: simple message, formatted message, and structured logging.
type Logger interface {
	// Debug logs messages for debugging purposes - typically very verbose
	Debug(message string)
	// Debugf logs formatted debug messages using a template string
	Debugf(template string, args ...any)
	// DebugS logs structured debug messages with additional options
	DebugS(template string, opts ...Option)

	// Info logs informational messages about application progress
	Info(message string)
	// Infof logs formatted info messages using a template string
	Infof(template string, args ...any)
	// InfoS logs structured info messages with additional options
	InfoS(template string, opts ...Option)

	// Warn logs warning messages about potential issues
	Warn(message string)
	// Warnf logs formatted warning messages using a template string
	Warnf(template string, args ...any)
	// WarnS logs structured warning messages with additional options
	WarnS(template string, opts ...Option)

	// Error logs error messages about failures that need attention
	Error(message string)
	// Errorf logs formatted error messages using a template string
	Errorf(template string, args ...any)
	// ErrorS logs structured error messages with additional options
	ErrorS(template string, opts ...Option)

	// Fatal logs critical errors and typically exits the application
	Fatal(message string)
	// Fatalf logs formatted fatal messages using a template string
	Fatalf(template string, args ...any)

	// With creates a new logger instance with additional options
	// Implementations should preserve existing configuration while adding new options
	With(opts ...Option) Logger
}

// Provider defines the interface for underlying logging implementations.
// It handles the actual writing of log messages and level checking.
type Provider interface {
	// Write outputs a log message with specified level and options
	Write(level Level, message string, opts ...Option)
	// WriteF outputs a formatted log message with specified level
	WriteF(level Level, template string, args ...any)
	// Enabled checks if a specific logging level is active
	Enabled(level Level) bool
}
