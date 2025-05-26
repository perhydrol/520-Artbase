package user

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *UserController) Get(c *gin.Context) {
	log.C(c).Infow("Get user", "userUUID", c.Param("useruuid"))
	useruuid := c.Param("useruuid")
	if !govalidator.IsUUIDv4(useruuid) {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}
	resp, err := ctrl.b.Users().Get(c, useruuid)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}
