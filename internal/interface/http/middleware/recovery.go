package middleware

import (
	"log"
	"runtime/debug"

	"im-chat/internal/interface/http/response"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\nStack trace: %s", err, debug.Stack())
				response.InternalError(c)
				c.Abort()
			}
		}()
		c.Next()
	}
}

func NoRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.NotFound(c)
	}
}

func NoMethod() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.FailedWithCodeAndMsg(c, response.CodeFailed, "method not allowed")
	}
}
