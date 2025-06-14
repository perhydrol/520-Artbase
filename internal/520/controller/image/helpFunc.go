package image

import (
	"demo520/internal/pkg/errno"
	"demo520/pkg/api"
	"demo520/pkg/token"

	"github.com/gin-gonic/gin"
)

func pareseJwtAndEqualReqUUID(ctx *gin.Context, req api.HasUserUUID) (jwtUserUUID string, err error) {
	reqUserUUID := ""
	if req != nil {
		reqUserUUID = req.GetUserUUID()
	}
	jwtUserUUID, err = token.ParseRequest(ctx)
	if err != nil {
		return "", errno.ErrTokenInvalid
	}
	if jwtUserUUID != reqUserUUID {
		return "", errno.ErrUnauthorized
	}
	return
}
