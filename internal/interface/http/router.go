package http

import (
	"im-chat/internal/interface/http/handler"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var userHandler = handler.NewUserHandler()

func RegisterRoutes(r *gin.Engine) {
	r.StaticFS("/public", http.Dir("public"))

	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join("public", "index.html"))
	})

	public := r.Group("/api")
	{
		public.POST("/user/register", userHandler.Register)
		public.POST("/user/login", userHandler.Login)
	}
}
