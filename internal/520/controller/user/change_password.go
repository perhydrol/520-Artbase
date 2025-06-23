package user

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/known"
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

	// 从JWT中获取用户邮箱
	email, exists := c.Get(known.XUsernameKey)
	if !exists {
		err := errno.ErrTokenInvalid
		log.ErrorWithFunc(err, "无法从JWT中获取用户信息")
		core.WriteResponse(c, err, nil)
		return
	}

	if err := ctrl.b.Users().ChangePassword(c.Request.Context(), email.(string), &r); err != nil {
		log.ErrorWithFunc(err, "修改密码失败", "email", email)
		core.WriteResponse(c, err, nil)
		return
	}

	log.Infow("密码修改成功", "email", email)
	core.WriteResponse(c, nil, nil)
}
