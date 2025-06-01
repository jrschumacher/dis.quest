package main

import (
	_ "embed"

	"github.com/jrschumacher/dis.quest/cmd"
	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/logger"
)

//go:embed "keys/jwks.public.json"
var jwksPublic string

func main() {
	// Load config with embedded JWKS
	cfg := config.Load(jwksPublic)
	logger.Init(cfg.LogLevel)

	cmd.Execute(cfg)
}
