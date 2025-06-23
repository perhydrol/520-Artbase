package user

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/pkg/api"

	"github.com/gin-gonic/gin"
)

func (ctrl *UserController) ChangePassword(c *gin.Context) {
	defer log.FuncEntryWithContext(c.Request.Context())()

	var r api.ChangePasswordRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		log.ErrorWithFunc(err, "请求参数绑定失败")
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	// 获取用户邮箱
	email := c.Param("email")

	if err := ctrl.b.Users().ChangePassword(c.Request.Context(), email, &r); err != nil {
		log.ErrorWithFunc(err, "修改密码失败", "email", email)
		core.WriteResponse(c, err, nil)
		return
	}

	log.Infow("密码修改成功", "email", email)
	core.WriteResponse(c, nil, nil)
}
