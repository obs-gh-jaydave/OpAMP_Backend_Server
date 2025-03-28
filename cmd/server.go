package main

import (
	"log"
	"opamp-backend/internal/server"
)

func main() {
	srv, err := server.NewServer("config/backend.yaml")
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	log.Println("OpAMP Backend Server running...")
	srv.Start()
}
