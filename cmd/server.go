package cmd

import (
	"github.com/jrschumacher/dis.quest/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"start"},
	Short:   "Start the dis.quest server",
	Run: func(_ *cobra.Command, _ []string) {
		server.Start(cfg)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
