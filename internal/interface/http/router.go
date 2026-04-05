package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	// Serve static files
	r.StaticFS("/public", http.Dir("public"))
	// Public routes
	public := r.Group("/api")
	{
		public.POST("/auth/register", RegisterHandler)
		public.POST("/auth/login", LoginHandler)
		public.POST("/auth/refresh", RefreshTokenHandler)
	}

	// Protected routes
	protected := r.Group("/api")
	protected.Use(JWTMiddleware())
	{
		// User routes
		user := protected.Group("/user")
		{
			user.GET("/profile", GetProfileHandler)
			user.PUT("/profile", UpdateProfileHandler)
		}

		// Friend routes
		friend := protected.Group("/friend")
		{
			friend.GET("/list", GetFriendListHandler)
			friend.POST("/add", AddFriendHandler)
			friend.POST("/delete", DeleteFriendHandler)
		}

		// Group routes
		group := protected.Group("/group")
		{
			group.GET("/list", GetGroupListHandler)
			group.POST("/create", CreateGroupHandler)
			group.POST("/join", JoinGroupHandler)
			group.POST("/leave", LeaveGroupHandler)
		}

		// Message routes
		message := protected.Group("/message")
		{
			message.GET("/history", GetMessageHistoryHandler)
		}
	}
}

// Handler stubs
func RegisterHandler(c *gin.Context)          { c.JSON(200, gin.H{"msg": "register"}) }
func LoginHandler(c *gin.Context)             { c.JSON(200, gin.H{"msg": "login"}) }
func RefreshTokenHandler(c *gin.Context)      { c.JSON(200, gin.H{"msg": "refresh"}) }
func JWTMiddleware() gin.HandlerFunc          { return func(c *gin.Context) {} }
func GetProfileHandler(c *gin.Context)        { c.JSON(200, gin.H{"msg": "profile"}) }
func UpdateProfileHandler(c *gin.Context)     { c.JSON(200, gin.H{"msg": "update profile"}) }
func GetFriendListHandler(c *gin.Context)     { c.JSON(200, gin.H{"msg": "friend list"}) }
func AddFriendHandler(c *gin.Context)         { c.JSON(200, gin.H{"msg": "add friend"}) }
func DeleteFriendHandler(c *gin.Context)      { c.JSON(200, gin.H{"msg": "delete friend"}) }
func GetGroupListHandler(c *gin.Context)      { c.JSON(200, gin.H{"msg": "group list"}) }
func CreateGroupHandler(c *gin.Context)       { c.JSON(200, gin.H{"msg": "create group"}) }
func JoinGroupHandler(c *gin.Context)         { c.JSON(200, gin.H{"msg": "join group"}) }
func LeaveGroupHandler(c *gin.Context)        { c.JSON(200, gin.H{"msg": "leave group"}) }
func GetMessageHistoryHandler(c *gin.Context) { c.JSON(200, gin.H{"msg": "message history"}) }
