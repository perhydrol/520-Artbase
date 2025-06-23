package middleware

import (
	"demo520/internal/pkg/core"
	"demo520/pkg/token"
	"errors"

	"github.com/gin-gonic/gin"
)

const CtxJWTUserUUID = "jwt-userUUID"

func JWTAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtUserUUID, err := token.ParseRequest(ctx)
		if err != nil {
			if errors.Is(err, token.ErrMissingHeader) {
				ctx.Set(CtxJWTUserUUID, "")
				ctx.Next()
			} else {
				core.WriteResponse(ctx, token.ErrInvalidToken, nil)
				ctx.Abort()
				return
			}
		}
		ctx.Set(CtxJWTUserUUID, jwtUserUUID)
		ctx.Next()
	}
}
