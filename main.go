// Package main is the entry point for the dis.quest application
package main

import (
	"github.com/jrschumacher/dis.quest/cmd"
	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/logger"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.LogLevel)

	cmd.Execute(cfg)
}
