package main

import (
	http_interface "im-chat/internal/interface/http"
	"im-chat/internal/interface/websocket"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Register HTTP routes
	http_interface.RegisterRoutes(r)

	// Register WebSocket server
	r.GET("/room", func(c *gin.Context) {
		websocket.WebSocketHandler(c.Writer, c.Request)
	})

	// Start server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
