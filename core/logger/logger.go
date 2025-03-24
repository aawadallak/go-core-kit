package logger

import (
	"os"
)

// logger implements the Logger interface, managing logging operations through a Provider
type logger struct {
	provider Provider // Underlying logging implementation
	cfg      []Option // Configuration options
}

// Compile-time check for Logger interface implementation
var _ Logger = (*logger)(nil)

// New creates a new logger instance with the specified options
func New(opts ...Option) *logger {
	l := &logger{
		provider: initConsoleLogger(), // Default provider
		cfg:      opts,                // Initial configuration
	}

	// If global logger exists, inherit its configuration
	if global != nil {
		if globalLogger, ok := global.logger.(*logger); ok {
			l.cfg = append(globalLogger.cfg, opts...)
		}
	}

	// Apply configuration to potentially override provider
	if cfg := NewConfig(l.cfg...); cfg.Provider != nil {
		l.provider = cfg.Provider
	}

	return l
}

// With creates a new logger instance with additional options appended to existing ones
func (l *logger) With(opts ...Option) Logger {
	// Create a new slice to avoid modifying the original
	newCfg := make([]Option, len(l.cfg))
	copy(newCfg, l.cfg)
	newCfg = append(newCfg, opts...)

	return &logger{
		provider: l.provider, // Keep same provider
		cfg:      newCfg,     // Use new configuration
	}
}

// Debug logs a debug message if enabled
func (l *logger) Debug(message string) {
	if l.provider.Enabled(DebugLevel) {
		l.provider.Write(DebugLevel, message, l.cfg...)
	}
}

// Debugf logs a formatted debug message if enabled
func (l *logger) Debugf(template string, args ...any) {
	if l.provider.Enabled(DebugLevel) {
		l.provider.WriteF(DebugLevel, template, args...)
	}
}

// DebugS logs a structured debug message with additional options
func (l *logger) DebugS(template string, opts ...Option) {
	if l.provider.Enabled(DebugLevel) {
		l.provider.Write(DebugLevel, template, append(l.cfg, opts...)...)
	}
}

// Info logs an info message if enabled
func (l *logger) Info(message string) {
	if l.provider.Enabled(InfoLevel) {
		l.provider.Write(InfoLevel, message, l.cfg...)
	}
}

// Infof logs a formatted info message if enabled
func (l *logger) Infof(template string, args ...any) {
	if l.provider.Enabled(InfoLevel) {
		l.provider.WriteF(InfoLevel, template, args...)
	}
}

// InfoS logs a structured info message with additional options
func (l *logger) InfoS(template string, opts ...Option) {
	if l.provider.Enabled(InfoLevel) {
		l.provider.Write(InfoLevel, template, append(l.cfg, opts...)...)
	}
}

// Warn logs a warning message if enabled
func (l *logger) Warn(message string) {
	if l.provider.Enabled(WarnLevel) {
		l.provider.Write(WarnLevel, message, l.cfg...)
	}
}

// Warnf logs a formatted warning message if enabled
func (l *logger) Warnf(template string, args ...any) {
	if l.provider.Enabled(WarnLevel) {
		l.provider.WriteF(WarnLevel, template, args...)
	}
}

// WarnS logs a structured warning message with additional options
func (l *logger) WarnS(template string, opts ...Option) {
	if l.provider.Enabled(WarnLevel) {
		l.provider.Write(WarnLevel, template, append(l.cfg, opts...)...)
	}
}

// Error logs an error message if enabled
func (l *logger) Error(message string) {
	if l.provider.Enabled(ErrorLevel) {
		l.provider.Write(ErrorLevel, message, l.cfg...)
	}
}

// Errorf logs a formatted error message if enabled
func (l *logger) Errorf(template string, args ...any) {
	if l.provider.Enabled(ErrorLevel) {
		l.provider.WriteF(ErrorLevel, template, args...)
	}
}

// ErrorS logs a structured error message with additional options
func (l *logger) ErrorS(template string, opts ...Option) {
	if l.provider.Enabled(ErrorLevel) {
		l.provider.Write(ErrorLevel, template, append(l.cfg, opts...)...)
	}
}

// Fatal logs a fatal message and exits if enabled
func (l *logger) Fatal(message string) {
	if l.provider.Enabled(FatalLevel) {
		l.provider.Write(FatalLevel, message, l.cfg...)
		os.Exit(1)
	}
}

// Fatalf logs a formatted fatal message and exits if enabled
func (l *logger) Fatalf(template string, args ...any) {
	if l.provider.Enabled(FatalLevel) {
		l.provider.WriteF(FatalLevel, template, args...)
		os.Exit(1)
	}
}
