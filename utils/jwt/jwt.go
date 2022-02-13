package jwt

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type JWT struct {
	SigningKey  []byte
	ExpiresTime int64
	BufferTime  int64
	Issuer      string
}

var (
	TokenExpired     = errors.New("Token is expired")
	TokenNotValidYet = errors.New("Token not active yet")
	TokenMalformed   = errors.New("That's not even a token")
	TokenInvalid     = errors.New("Couldn't handle this token:")
)

func NewJWT(config JWTConfig) *JWT {
	return &JWT{
		SigningKey:  []byte(config.SigningKey),
		ExpiresTime: config.ExpiresTime,
		BufferTime:  config.BufferTime,
		Issuer:      config.Issuer,
	}
}

func (j *JWT) CreateClaims(baseClaims JWTBaseClaims) JWTCustomClaims {
	claims := JWTCustomClaims{
		JWTBaseClaims: baseClaims,
		BufferTime:    j.BufferTime, // 缓冲时间1天 缓冲时间内会获得新的token刷新令牌 此时一个用户会存在两个有效令牌 但是前端只留一个 另一个会丢失
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix() - 1000,          // 签名生效时间
			ExpiresAt: time.Now().Unix() + j.ExpiresTime, // 过期时间 7天  配置文件
			Issuer:    j.Issuer,                          // 签名的发行者
		},
	}
	return claims
}

// 创建一个token
func (j *JWT) CreateToken(claims JWTCustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// 解析 token
func (j *JWT) ParseToken(tokenString string) (*JWTCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTCustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*JWTCustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, TokenInvalid

	} else {
		return nil, TokenInvalid

	}

}
