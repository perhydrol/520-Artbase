package middleware

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/log"
	"demo520/pkg/token"
	"errors"

	"github.com/gin-gonic/gin"
)

const CtxJWTUserUUID = "jwt-userUUID"

func JWTAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.C(ctx).Infow("JWT authentication started",
			"path", ctx.Request.URL.Path,
			"method", ctx.Request.Method,
			"client_ip", ctx.ClientIP(),
			"user_agent", ctx.Request.UserAgent(),
		)

		jwtUserUUID, err := token.ParseRequest(ctx)
		if err != nil {
			if errors.Is(err, token.ErrMissingHeader) {
				log.C(ctx).Infow("JWT token missing, allowing anonymous access",
					"path", ctx.Request.URL.Path,
					"method", ctx.Request.Method,
					"client_ip", ctx.ClientIP(),
				)
				ctx.Set(CtxJWTUserUUID, "")
				ctx.Next()
			} else {
				log.C(ctx).Warnw("JWT authentication failed",
					"error", err.Error(),
					"path", ctx.Request.URL.Path,
					"method", ctx.Request.Method,
					"client_ip", ctx.ClientIP(),
					"user_agent", ctx.Request.UserAgent(),
				)
				core.WriteResponse(ctx, token.ErrInvalidToken, nil)
				ctx.Abort()
				return
			}
		} else {
			log.C(ctx).Infow("JWT authentication successful",
				"user_uuid", jwtUserUUID,
				"path", ctx.Request.URL.Path,
				"method", ctx.Request.Method,
				"client_ip", ctx.ClientIP(),
			)
			ctx.Set(CtxJWTUserUUID, jwtUserUUID)
		}
		ctx.Next()
	}
}
