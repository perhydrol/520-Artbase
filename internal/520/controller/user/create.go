package user

import (
	"demo520/internal/pkg/log"
	"demo520/pkg/api"
	"github.com/gin-gonic/gin"
)

func (ctrl *UserController) Create(c *gin.Context) {
	log.C(c).Infow("create user")

	var r api.CreateUserRequest
	if err := c.ShouldBindJSON(&r); err != nil {

	}
}
