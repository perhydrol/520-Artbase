package user

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/pkg/api"

	"github.com/gin-gonic/gin"
)

func (ctrl *UserController) ChangePassword(c *gin.Context) {
	defer log.C(c).FuncEntryWithContext(c, "email", "[REDACTED]")() // 不记录敏感信息

	// 获取用户邮箱
	email := c.Param("email")

	log.C(c).Infow("Password change attempt started",
		"email", email,
		"client_ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
	)

	var r api.ChangePasswordRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		log.C(c).Errorw("Failed to bind password change request",
			"error", err.Error(),
			"email", email,
			"client_ip", c.ClientIP(),
		)
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	log.C(c).Infow("Password change request validation started",
		"email", email,
		"has_old_password", r.OldPassword != "",
		"has_new_password", r.NewPassword != "",
	)

	log.C(c).Infow("Calling business layer for password change",
		"email", email,
	)

	if err := ctrl.b.Users().ChangePassword(c.Request.Context(), email, &r); err != nil {
		log.C(c).Errorw("Password change failed",
			"error", err.Error(),
			"email", email,
			"client_ip", c.ClientIP(),
		)
		core.WriteResponse(c, err, nil)
		return
	}

	log.C(c).Infow("Password change successful",
		"email", email,
		"client_ip", c.ClientIP(),
	)
	core.WriteResponse(c, nil, nil)
}
