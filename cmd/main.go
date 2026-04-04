package main

import (
	"im-chat/internal/interface/http"
	"im-chat/internal/interface/socket"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Register HTTP routes
	http.RegisterRoutes(r)

	// Register Socket.IO server
	socketServer := socket.NewSocketServer()
	r.GET("/socket.io/*any", gin.WrapH(socketServer))
	r.POST("/socket.io/*any", gin.WrapH(socketServer))

	// Start server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
