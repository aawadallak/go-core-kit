package slogx

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aawadallak/go-core-kit/core/logger"
)

// SlogProvider is a logger provider that uses slog for logging.
type SlogProvider struct {
	logger *slog.Logger
	level  logger.Level
}

// Ensure SlogProvider implements the Provider interface.
var _ logger.Provider = (*SlogProvider)(nil)

// NewProvider creates a new SlogProvider with default configuration.
func NewProvider(opts ...Option) (logger.Provider, error) {
	options := newOptions(opts...)

	// Set the log level based on the environment variable
	logLevel := logger.DefaultLevel()
	var slogLevel slog.Level

	switch logLevel {
	case logger.DebugLevel:
		slogLevel = slog.LevelDebug
	case logger.InfoLevel:
		slogLevel = slog.LevelInfo
	case logger.WarnLevel:
		slogLevel = slog.LevelWarn
	case logger.ErrorLevel:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo // Fallback to InfoLevel
	}

	lOptions := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: false,
	}

	if options.addSource {
		lOptions.AddSource = true
	}

	// Configure slog handler with JSON output and level
	handler := slog.NewJSONHandler(os.Stdout, lOptions)
	logger := slog.New(handler)

	slog.SetDefault(logger)

	return &SlogProvider{
		logger: logger,
		level:  logLevel,
	}, nil
}

// Write logs a message with the specified level and attributes.
func (p *SlogProvider) Write(level logger.Level, message string, opts ...logger.Option) {
	if !p.Enabled(level) {
		return
	}

	cfg := logger.NewConfig(opts...)

	// Prepare attributes for structured logging
	// We'll pass these as variadic arguments instead of a slice
	args := make([]any, 0, len(cfg.Attributes)*2) // *2 for key-value pairs
	for k, v := range cfg.Attributes {
		args = append(args, k, v)
	}

	switch level {
	case logger.InfoLevel:
		p.logger.Info(message, args...)
	case logger.WarnLevel:
		p.logger.Warn(message, args...)
	case logger.ErrorLevel:
		p.logger.Error(message, args...)
	case logger.DebugLevel:
		p.logger.Debug(message, args...)
	default:
		p.logger.Info(message, args...) // Fallback to Info
	}
}

// WriteF formats and logs a message with the specified level.
func (p *SlogProvider) WriteF(level logger.Level, template string, args ...any) {
	if !p.Enabled(level) {
		return
	}
	msg := fmt.Sprintf(template, args...)
	p.Write(level, msg)
}

// Enabled checks if a certain logging level is enabled.
func (p *SlogProvider) Enabled(level logger.Level) bool {
	return level >= p.level
}
