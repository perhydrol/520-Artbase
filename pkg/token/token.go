package token

import (
	"demo520/internal/pkg/log"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"sync"
	"time"
)

var ErrMissingHeader = errors.New("the length of the `Authorization` header is zero")
var ErrInvalidToken = errors.New("the `Authorization` header is invalid")
var ErrSigningMethod = errors.New("the `Authorization` signing method is invalid")

var (
	jwtSecretKey string
	once         sync.Once
)

type CustomClaims struct {
	UserUUID string `json:"useruuid" valid:"required,uuidv4"`
	jwt.RegisteredClaims
}

func Init(key string) {
	once.Do(func() {
		if key == "" {
			log.Warnw("Using default key")
			key = "wK3NpsaF0LsjkXagIelqiHWbaKKjp48rqAcK8lPXvrRELBRKi4Gthfjqqx8BH9jW"
		}
		jwtSecretKey = key
	})
}

func GenerateToken(userUUID string) (string, error) {
	claims := CustomClaims{
		UserUUID: userUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecretKey))
}

func ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrSigningMethod
		}
		return []byte(jwtSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ParseRequest 从请求头中获取令牌，并将其传递给 Parse 函数以解析令牌.
func ParseRequest(c *gin.Context) (string, error) {
	header := c.Request.Header.Get("Authorization")

	if len(header) == 0 {
		return "", ErrMissingHeader
	}
	// 从请求头中取出 token
	if !strings.HasPrefix(header, "Bearer ") {
		return "", ErrInvalidToken
	}

	t := strings.TrimPrefix(header, "Bearer ")

	claims, err := ParseToken(t)
	return claims.UserUUID, err
}
