package main

import (
	"github.com/aawadallak/go-core-kit/core/logger"
	"github.com/aawadallak/go-core-kit/plugin/logger/zapx"
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	lProvider, err := zapx.NewProvider()
	handleError(err)

	l := logger.New(logger.WithProvider(lProvider))
	l.Info("Hello, World!")
}
