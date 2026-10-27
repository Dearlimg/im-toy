package main

import (
	"fmt"
	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/config"
	"im-chat/internal/infra/database"
	"im-chat/internal/infra/jwt"
	"im-chat/internal/infra/redis"
	http_interface "im-chat/internal/interface/http"
	"im-chat/internal/interface/websocket"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gin-gonic/gin"
)

func findConfigPath() string {
	possiblePaths := []string{
		filepath.Join("config", "config.yaml"),
		filepath.Join("..", "config", "config.yaml"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return filepath.Join("config", "config.yaml")
}

func main() {
	cfgPath := findConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.InitDB(&cfg.Database); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	if err := database.DB.AutoMigrate(
		&entity.User{},
		&entity.Friend{},
		&entity.PrivateMessage{},
		&entity.Group{},
		&entity.GroupMember{},
		&entity.GroupMessage{},
		&entity.UserRoomRead{},
		&entity.MessageRead{},
		&entity.MessageRequest{},
		&entity.MessageConversation{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	// Old rows must compare as zero rather than SQL NULL in cursor queries.
	for _, table := range []string{"private_messages", "group_messages"} {
		if err := database.DB.Exec("UPDATE " + table + " SET seq = 0 WHERE seq IS NULL").Error; err != nil {
			log.Fatalf("Failed to initialize legacy message sequences: %v", err)
		}
	}

	if err := redis.InitRedis(&cfg.Redis); err != nil {
		log.Printf("Warning: Failed to init redis: %v", err)
	}

	jwt.InitJWT(&cfg.JWT)

	log.Println("Messages use MySQL and local push; legacy Kafka consumers are disabled")

	r := gin.Default()
	http_interface.RegisterRoutes(r)

	r.GET("/room", func(c *gin.Context) {
		websocket.WebSocketHandler(c.Writer, c.Request)
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("Server starting on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	log.Println("Server exited")
}
