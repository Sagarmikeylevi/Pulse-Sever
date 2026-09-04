package commands

import (
	"log"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	Run: func(cmd *cobra.Command, args []string) {
		if err := shared.MigrateUp(cfg.DB); err != nil {
			log.Fatal(err)
		}
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration",
	Run: func(cmd *cobra.Command, args []string) {
		if err := shared.MigrateDown(cfg.DB); err != nil {
			log.Fatal(err)
		}
	},
}

var migrateDownAllCmd = &cobra.Command{
	Use:   "down-all",
	Short: "Rollback all migrations",
	Run: func(cmd *cobra.Command, args []string) {
		if err := shared.MigrateDownAll(cfg.DB); err != nil {
			log.Fatal(err)
		}
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new migration file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := shared.MigrateCreate(args[0]); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateDownAllCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	rootCmd.AddCommand(migrateCmd)
}
