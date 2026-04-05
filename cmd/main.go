package main

import (
	"fmt"
	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/config"
	"im-chat/internal/infra/database"
	"im-chat/internal/infra/jwt"
	"im-chat/internal/infra/kafka"
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
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	if err := redis.InitRedis(&cfg.Redis); err != nil {
		log.Printf("Warning: Failed to init redis: %v", err)
	}

	jwt.InitJWT(&cfg.JWT)

	kafka.InitTopics(&cfg.Kafka)

	if err := kafka.InitProducer(&cfg.Kafka); err != nil {
		log.Printf("Warning: Failed to init kafka producer: %v", err)
	}

	persistConsumer, err := kafka.NewPersistConsumer(
		cfg.Kafka.Brokers,
		"persist-consumer-group",
		[]string{kafka.TopicGroupMsgRaw, kafka.TopicPrivateMsg, kafka.TopicReadEvents},
	)
	if err != nil {
		log.Printf("Warning: Failed to create persist consumer: %v", err)
	} else {
		persistConsumer.Start()
	}

	pusherConsumer, err := kafka.NewPusherConsumer(
		cfg.Kafka.Brokers,
		"pusher-consumer-group",
		[]string{kafka.TopicGroupMsgPublish, kafka.TopicPrivateMsg},
	)
	if err != nil {
		log.Printf("Warning: Failed to create pusher consumer: %v", err)
	} else {
		pusherConsumer.Start()
	}

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

	if persistConsumer != nil {
		persistConsumer.Stop()
	}
	if pusherConsumer != nil {
		pusherConsumer.Stop()
	}
	kafka.Close()

	log.Println("Server exited")
}
