package http

import (
	"im-chat/internal/application/service/impl"
	infra_repo "im-chat/internal/infra/repository"
	"im-chat/internal/interface/http/handler"
	"im-chat/internal/interface/http/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	userRepo := infra_repo.NewUserRepository()
	friendRepo := infra_repo.NewFriendRepository()
	groupRepo := infra_repo.NewGroupRepository()

	userService := impl.NewUserServiceImpl(userRepo)
	friendService := impl.NewFriendServiceImpl(friendRepo)
	groupService := impl.NewGroupServiceImpl(groupRepo)

	userHandler := handler.NewUserHandler(userService)
	friendHandler := handler.NewFriendHandler(friendService)
	groupHandler := handler.NewGroupHandler(groupService)
	privateMsgHandler := handler.NewPrivateMessageHandler()

	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	r.StaticFS("/public", http.Dir("public"))

	// Use redirect so "/" always hits the same static file pipeline as /public/* (avoids 404 when cwd
	// makes c.File("public/index.html") miss while StaticFS still resolves "public" correctly).
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/public/index.html")
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

		auth.POST("/group/create", groupHandler.CreateGroup)
		auth.GET("/group/list", groupHandler.GetGroupList)
		auth.GET("/group/members", groupHandler.GetGroupMembers)
		auth.GET("/group/messages", groupHandler.GetGroupMessages)
	}

	r.NoRoute(middleware.NoRoute())
	r.NoMethod(middleware.NoMethod())
}
