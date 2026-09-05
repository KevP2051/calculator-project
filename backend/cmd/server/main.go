package main

import (
	"log"
	"net/http"
	"os"

	"calculator/backend/internal/api"
)

func main() {
	address := ":" + envOr("PORT", "8080")
	handler := api.NewRouter(envOr("CORS_ORIGIN", "http://localhost:5173"))

	log.Printf("calculator backend listening on %s", address)

	if err := http.ListenAndServe(address, handler); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
