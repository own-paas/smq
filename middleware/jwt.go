package middleware

import (
	"github.com/gin-gonic/gin"
	restful "github.com/sestack/grf"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/utils/jwt"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 我们这里jwt鉴权取头部信息 x-token 登录时回返回token信息 这里前端需要把token存储到cookie或者本地localStorage中 不过需要跟后端协商过期时间 可以约定刷新令牌或者重新登录
		token := c.Request.Header.Get("x-token")
		if token == "" {
			restful.Unauthorized(c)
			c.Abort()
			return
		}

		j := jwt.NewJWT(global.CONFIG.JWT)
		// parseToken 解析token包含的信息
		claims, err := j.ParseToken(token)
		if err != nil {
			if err == jwt.TokenExpired {
				restful.NotForbidden(c)
				c.Abort()
				return
			}
			restful.NotForbidden(c)
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
