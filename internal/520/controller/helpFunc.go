package controller

import (
	"demo520/internal/pkg/middleware"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

// GetUserUUIDFromContext 从 gin.Context 中提取 jwt-userUUID。
func GetUserUUIDFromContext(c *gin.Context) (string, bool) {
	v, exists := c.Get(middleware.CtxJWTUserUUID)
	if !exists {
		return "", false
	}
	uuid, ok := v.(string)
	if govalidator.IsUUIDv4(uuid) {
		return uuid, ok
	} else {
		return "", false
	}
}
