package user

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/pkg/api"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *UserController) Create(c *gin.Context) {
	defer log.C(c).FuncEntryWithContext(c, "email", "[REDACTED]")() // 不记录敏感信息

	log.C(c).Infow("User registration attempt started",
		"client_ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
	)

	var r api.CreateUserRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		log.C(c).Errorw("Failed to bind user creation request",
			"error", err.Error(),
			"client_ip", c.ClientIP(),
		)
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	log.C(c).Infow("User creation request validation started",
		"email", r.Email,
		"nickname", r.Nickname,
		"has_password", r.Password != "",
	)

	if _, err := govalidator.ValidateStruct(r); err != nil {
		log.C(c).Errorw("User creation request validation failed",
			"error", err.Error(),
			"email", r.Email,
			"nickname", r.Nickname,
		)
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}

	log.C(c).Infow("Calling business layer for user creation",
		"email", r.Email,
		"nickname", r.Nickname,
	)

	userInfo, err := ctrl.b.Users().Create(c, &r)
	if err != nil {
		log.C(c).Errorw("User creation failed",
			"error", err.Error(),
			"email", r.Email,
			"nickname", r.Nickname,
			"client_ip", c.ClientIP(),
		)
		core.WriteResponse(c, err, nil)
		return
	}

	log.C(c).Infow("User creation successful",
		"email", r.Email,
		"nickname", r.Nickname,
		"user_uuid", userInfo.UserUUID,
		"client_ip", c.ClientIP(),
	)

	core.WriteResponse(c, nil, userInfo)
}
