package http

import (
	"im-chat/internal/di"
	"im-chat/internal/interface/http/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	container := di.NewContainer()

	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	r.StaticFS("/public", http.Dir("public"))

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/public/index.html")
	})

	public := r.Group("/api")
	{
		public.POST("/user/register", container.UserHandler.Register)
		public.POST("/user/login", container.UserHandler.Login)
	}

	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/friend/search", container.FriendHandler.SearchUser)
		auth.POST("/friend/add", container.FriendHandler.AddFriend)
		auth.POST("/friend/accept", container.FriendHandler.AcceptFriend)
		auth.GET("/friend/list", container.FriendHandler.GetFriendList)
		auth.GET("/friend/requests", container.FriendHandler.GetFriendRequests)
		auth.POST("/friend/delete", container.FriendHandler.DeleteFriend)

		auth.POST("/private/send", container.PrivateMsgHandler.SendMessage)
		auth.GET("/private/conversation", container.PrivateMsgHandler.GetConversation)
		auth.GET("/private/unread", container.PrivateMsgHandler.GetUnreadCount)
		auth.POST("/private/mark-read", container.PrivateMsgHandler.MarkAsRead)

		auth.POST("/group/create", container.GroupHandler.CreateGroup)
		auth.GET("/group/list", container.GroupHandler.GetGroupList)
		auth.GET("/group/members", container.GroupHandler.GetGroupMembers)
		auth.GET("/group/messages", container.GroupHandler.GetGroupMessages)

		auth.POST("/message/report-read", container.MessageHandler.ReportRead)
		auth.GET("/message/unread-count", container.MessageHandler.GetUnreadCount)
		auth.GET("/message/read-status", container.MessageHandler.GetRoomReadStatus)
	}

	r.NoRoute(middleware.NoRoute())
	r.NoMethod(middleware.NoMethod())
}
