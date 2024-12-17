package routes

import (
	"pubsub-server/controllers"
	"pubsub-server/middleware"

	"github.com/gorilla/mux"
)

// InitRoutes initializes all routes
func InitRoutes() *mux.Router {
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.Logging)

	// Pub/Sub Routes
	router.HandleFunc("/publish", controllers.Publish).Methods("POST")
	router.HandleFunc("/subscribe", controllers.Subscribe).Methods("POST")
	router.HandleFunc("/subscribessl", controllers.SubscribeSSE).Methods("GET")

	return router
}
