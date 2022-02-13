package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type JWTCustomClaims struct {
	JWTBaseClaims
	BufferTime int64
	jwt.StandardClaims
}

type JWTBaseClaims struct {
	Username string
	Role     uint
}

func GetClaims(c *gin.Context, config JWTConfig) (*JWTCustomClaims, error) {
	token := c.Request.Header.Get("x-token")
	j := NewJWT(config)
	claims, err := j.ParseToken(token)

	return claims, err
}

func GetUserRole(c *gin.Context, config JWTConfig) uint {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c, config); err != nil {
			return 0
		} else {
			return cl.Role
		}
	} else {
		waitUse := claims.(*JWTCustomClaims)
		return waitUse.Role
	}

}

func GetUserName(c *gin.Context, config JWTConfig) string {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c, config); err != nil {
			return ""
		} else {
			return cl.Username
		}
	} else {
		waitUse := claims.(*JWTCustomClaims)
		return waitUse.Username
	}

}
