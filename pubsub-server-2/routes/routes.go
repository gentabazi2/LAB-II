// routes/routes.go
package routes

import (
	"pubsub-server/controllers"
	"pubsub-server/middleware"

	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.Logging)
	router.Use(middleware.CORS)

	// Topic management
	router.HandleFunc("/topics", controllers.GetTopics).Methods("GET", "OPTIONS")
	router.HandleFunc("/subscribe", controllers.Subscribe).Methods("POST", "OPTIONS")
	router.HandleFunc("/publish", controllers.Publish).Methods("POST", "OPTIONS")

	// Connection endpoints
	router.HandleFunc("/subscribe/sse", controllers.SubscribeSSE).Methods("GET", "OPTIONS")
	router.HandleFunc("/subscribe/ws", controllers.SubscribeWS).Methods("GET", "OPTIONS")

	return router
}
