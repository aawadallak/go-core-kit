package zapx

import (
	"fmt"

	"github.com/aawadallak/go-core-kit/core/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapProvider is a logger provider that uses Zap for logging.
type ZapProvider struct {
	logger  *zap.Logger
	level   logger.Level
	options *options
}

// Ensure ZapProvider implements the Provider interface.
var _ logger.Provider = (*ZapProvider)(nil)

// defaultZapProvider creates a new ZapProvider with the default configuration.
func NewProvider(opts ...Option) (logger.Provider, error) {
	options := newOptions(opts...)

	cfg := zap.NewProductionConfig()

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	cfg.DisableStacktrace = true
	if options.enableStackTrace {
		cfg.DisableStacktrace = false
	}

	cfg.DisableCaller = true
	if options.enableCaller {
		cfg.DisableCaller = false
	}

	// Set the log level based on the environment variable
	logLevel := logger.DefaultLevel()
	switch logLevel {
	case logger.DebugLevel:
		cfg.Level.SetLevel(zapcore.DebugLevel)
	case logger.InfoLevel:
		cfg.Level.SetLevel(zapcore.InfoLevel)
	case logger.WarnLevel:
		cfg.Level.SetLevel(zapcore.WarnLevel)
	case logger.ErrorLevel:
		cfg.Level.SetLevel(zapcore.ErrorLevel)
	default:
		cfg.Level.SetLevel(zapcore.InfoLevel) // Fallback to InfoLevel
	}

	lOptions := []zap.Option{}

	logger, err := cfg.Build(lOptions...)
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(logger)

	return &ZapProvider{logger: logger, level: logLevel, options: options}, nil
}

// Write logs a message with the specified level and attributes.
func (p *ZapProvider) Write(level logger.Level, message string, opts ...logger.Option) {
	if !p.Enabled(level) {
		return
	}

	cfg := logger.NewConfig(opts...)

	// Prepare attributes for structured logging
	fields := make([]zap.Field, 0, len(cfg.Attributes)+1)
	for k, v := range cfg.Attributes {
		fields = append(fields, zap.Any(k, v))
	}

	l := p.logger
	if p.options.enableCaller {
		l = p.logger.WithOptions(zap.AddCallerSkip(2))
	}

	switch level {
	case logger.InfoLevel:
		l.Info(message, fields...)
	case logger.WarnLevel:
		l.Warn(message, fields...)
	case logger.ErrorLevel:
		l.Error(message, fields...)
	case logger.DebugLevel:
		l.Debug(message, fields...)
	default:
		l.Info(message, fields...)
	}
}

// WriteF formats and logs a message with the specified level.
func (p *ZapProvider) WriteF(level logger.Level, template string, args ...any) {
	if !p.Enabled(level) {
		return
	}
	msg := fmt.Sprintf(template, args...)
	p.Write(level, msg)
}

// Enabled checks if a certain logging level is enabled.
func (p *ZapProvider) Enabled(level logger.Level) bool {
	return level >= p.level
}
