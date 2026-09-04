package commands

import (
	"log"
	"os"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
	"github.com/spf13/cobra"
)

var cfg *shared.Config

var rootCmd = &cobra.Command{
	Use:   "pulse",
	Short: "Pulse API server",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(loadConfig)
}

func loadConfig() {
	var err error
	cfg, err = shared.LoadConfig()
	if err != nil {
		log.Fatalf("unable to parse env vars: %v", err)
	}
}
