package user

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *UserController) Get(c *gin.Context) {
	log.C(c).Infow("Get user", "email", c.Param("email"))
	email := c.Param("email")
	if !govalidator.IsEmail(email) {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}
	resp, err := ctrl.b.Users().Get(c, email)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}
