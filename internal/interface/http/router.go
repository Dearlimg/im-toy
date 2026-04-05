package http

import (
	"im-chat/internal/interface/http/handler"
	"im-chat/internal/interface/http/middleware"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var (
	userHandler       = handler.NewUserHandler()
	friendHandler     = handler.NewFriendHandler()
	privateMsgHandler = handler.NewPrivateMessageHandler()
)

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

	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/friend/search", friendHandler.SearchUser)
		auth.POST("/friend/add", friendHandler.AddFriend)
		auth.POST("/friend/accept", friendHandler.AcceptFriend)
		auth.GET("/friend/list", friendHandler.GetFriendList)
		auth.GET("/friend/requests", friendHandler.GetFriendRequests)
		auth.POST("/friend/delete", friendHandler.DeleteFriend)

		auth.POST("/private/send", privateMsgHandler.SendMessage)
		auth.GET("/private/conversation", privateMsgHandler.GetConversation)
		auth.GET("/private/unread", privateMsgHandler.GetUnreadCount)
		auth.POST("/private/mark-read", privateMsgHandler.MarkAsRead)
	}
}
