package cmd

import (
	"os"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "disquest",
	Short: "dis.quest CLI",
	Long:  `dis.quest — Go POC for ATProtocol Discussions`,
}

func Execute(c *config.Config) {
	cfg = c
	logger.Info("Starting CLI", "env", cfg.AppEnv)
	if err := rootCmd.Execute(); err != nil {
		logger.Error("CLI error", "error", err)
		os.Exit(1)
	}
}
