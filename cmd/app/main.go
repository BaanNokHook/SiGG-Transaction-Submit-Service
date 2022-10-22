package main

import (
	"log"
	"nextclan/transaction-gateway/transaction-submit-service/config"
	"nextclan/transaction-gateway/transaction-submit-service/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
