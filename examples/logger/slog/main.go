package main

import (
	"context"

	"github.com/aawadallak/go-core-kit/core/logger"
	"github.com/aawadallak/go-core-kit/plugin/logger/slogx"
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// Create a background context as the base context
	ctx := context.Background()

	// Initialize a new slogx logging provider
	lProvider, err := slogx.NewProvider()
	// Check for any errors during provider creation
	handleError(err)

	// Create a new logger instance with the slogx provider
	l := logger.New(logger.WithProvider(lProvider))
	// Log an info message using the base logger
	l.Info("Hello, World!")

	// Set the logger as the global instance
	logger.SetInstance(l)
	// Add the logger to the context with an additional key-value pair
	ctx = logger.WithContext(ctx, logger.WithValue("ctxKey", "ctxValue"))
	// Get logger from context and log a warning message
	logger.Of(ctx).Warn("Hello, World!")

	// Create a new logger with an additional key-value pair based on the first logger
	l2 := l.With(logger.WithValue("key", "value"))
	// Log an info message with the new logger
	l2.Info("Hello, World!")

	// Create another logger building on l2 with another key-value pair
	l3 := l2.With(logger.WithValue("key2", "value2"))
	// Log an info message with all accumulated context
	l3.Info("Hello, World!")

	// Set l3 as the new global logger instance
	logger.SetInstance(l3)
	// Log an info message using the global logger
	logger.Global().Info("Hello, World!")
}
