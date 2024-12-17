package main

import (
	"log"
	"net/http"
	"pubsub-server/routes"
)

func main() {
	// Initialize routes
	router := routes.InitRoutes()

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("Server is running on port 8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
