package main

import (
	"log"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
)

func main() {
	cfg, err := shared.LoadConfig()

	if err != nil {
		log.Fatalf("unable to parse env vars: %v", err)
	}

	db, err := shared.NewDatabase(cfg.DB)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	log.Println("postgres connected")
}
