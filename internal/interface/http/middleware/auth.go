package middleware

import (
	"im-chat/internal/infra/jwt"
	"im-chat/internal/interface/http/response"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Failed(c, "未提供认证token")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Failed(c, "认证格式错误")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			response.Failed(c, "无效的token")
			c.Abort()
			return
		}

		c.Set("uid", claims.UID)
		c.Next()
	}
}
