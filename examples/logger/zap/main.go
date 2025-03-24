package main

import (
	"context"

	"github.com/aawadallak/go-core-kit/core/logger"
	"github.com/aawadallak/go-core-kit/plugin/logger/zapx"
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// Initialize a basic background context
	ctx := context.Background()

	// Create a new zapx logging provider instance
	lProvider, err := zapx.NewProvider()
	// Handle any potential errors from provider creation
	handleError(err)

	// Initialize a new logger with the zapx provider
	l := logger.New(logger.WithProvider(lProvider))
	// Log a basic info message using the initial logger
	l.Info("Hello, World!")

	// Set the logger as the global instance for the application
	logger.SetInstance(l)
	// Attach the logger to the context with an additional key-value pair
	ctx = logger.WithContext(ctx, logger.WithValue("ctxKey", "ctxValue"))
	// Retrieve logger from context and log a warning message
	logger.Of(ctx).Warn("Hello, World!")

	// Create a new logger instance with additional context based on the first logger
	l2 := l.With(logger.WithValue("key", "value"))
	// Log an info message with the updated context
	l2.Info("Hello, World!")

	// Create another logger layer with more context based on l2
	l3 := l2.With(logger.WithValue("key2", "value2"))
	// Log an info message with all accumulated context
	l3.Info("Hello, World!")

	// Update the global logger instance to use l3
	logger.SetInstance(l3)
	// Log an info message using the global logger with all context
	logger.Global().Info("Hello, World!")
}
