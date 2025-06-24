package user

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/pkg/api"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *UserController) Login(c *gin.Context) {
	defer log.C(c).FuncEntryWithContext(c, "email", "[REDACTED]")() // 不记录敏感信息

	log.C(c).Infow("User login attempt started",
		"client_ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
	)

	var r api.LoginRequest

	if err := c.ShouldBindJSON(&r); err != nil {
		log.C(c).Errorw("Failed to bind login request",
			"error", err.Error(),
			"client_ip", c.ClientIP(),
		)
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	log.C(c).Infow("Login request validation started",
		"email", r.Email, // 邮箱不是敏感信息，可以记录
		"has_password", r.Password != "",
		"seed_time", r.SeedTime,
	)

	if _, err := govalidator.ValidateStruct(r); err != nil {
		log.C(c).Errorw("Login request validation failed",
			"error", err.Error(),
			"email", r.Email,
		)
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}

	log.C(c).Infow("Calling business layer for login", "email", r.Email)

	resp, err := ctrl.b.Users().Login(c, &r)

	if err != nil {
		log.C(c).Errorw("Login failed",
			"error", err.Error(),
			"email", r.Email,
			"client_ip", c.ClientIP(),
		)
		core.WriteResponse(c, err, nil)
		return
	}

	log.C(c).Infow("Login successful",
		"email", r.Email,
		"client_ip", c.ClientIP(),
		"has_token", resp.Token != "",
	)

	core.WriteResponse(c, nil, resp)
}
