package user

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/pkg/api"
	"demo520/pkg/token"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *UserController) Update(c *gin.Context) {
	log.C(c).Infow("update user")

	var r api.UpdateUserRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	if _, err := govalidator.ValidateStruct(r); err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}

	userUUID, err := token.ParseRequest(c)
	if err != nil {
		core.WriteResponse(c, errno.ErrTokenInvalid, nil)
		return
	}
	if err := ctrl.b.Users().Update(c, userUUID, c.Param("email"), &r); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)
}
