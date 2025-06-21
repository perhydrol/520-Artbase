package middleware

import (
	"demo520/internal/pkg/core"
	"demo520/pkg/token"

	"github.com/gin-gonic/gin"
)

const CtxJWTUserUUID = "jwt-userUUID"

func JWTAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtUserUUID, err := token.ParseRequest(ctx)
		if err != nil {
			core.WriteResponse(ctx, err, nil)
			ctx.Abort()
			return
		}
		ctx.Set(CtxJWTUserUUID, jwtUserUUID)
		ctx.Next()
	}
}
