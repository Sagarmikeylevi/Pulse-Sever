package commands

import (
	"log"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/registry"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := shared.NewDatabase(cfg.DB)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			sqlDB, _ := db.DB()
			sqlDB.Close()
		}()

		log.Println("postgres connected")

		if err := shared.MigrateUp(cfg.DB); err != nil {
			log.Fatalf("migration error: %v", err)
		}

		router := registry.Setup(db, cfg)

		log.Printf("server starting on port %s", cfg.App.Port)
		if err := router.Run(":" + cfg.App.Port); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
