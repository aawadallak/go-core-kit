package main

import (
	"context"
	"fmt"

	"github.com/aawadallak/go-core-kit/core/conf"
	"github.com/aawadallak/go-core-kit/core/plugin/conf/ssm"
)

func main() {
	cfg := conf.New(context.Background(), conf.WithProvider(ssm.NewProvider()))
	conf.SetInstance(cfg)

	fmt.Println(cfg.GetString("ENV"))
	fmt.Println(cfg.GetString("DB_PASS"))
}
