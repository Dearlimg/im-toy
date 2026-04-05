package http

import (
	"im-chat/internal/application/service/impl"
	infra_repo "im-chat/internal/infra/repository"
	"im-chat/internal/interface/http/handler"
	"im-chat/internal/interface/http/middleware"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	userRepo := infra_repo.NewUserRepository()
	friendRepo := infra_repo.NewFriendRepository()

	userService := impl.NewUserServiceImpl(userRepo)
	friendService := impl.NewFriendServiceImpl(friendRepo)

	userHandler := handler.NewUserHandler(userService)
	friendHandler := handler.NewFriendHandler(friendService)
	privateMsgHandler := handler.NewPrivateMessageHandler()

	r.Use(middleware.Recovery())

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

	r.NoRoute(middleware.NoRoute())
	r.NoMethod(middleware.NoMethod())
}
