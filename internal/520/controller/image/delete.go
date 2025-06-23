package image

import (
	"demo520/internal/520/controller"
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) DeleteImage(ctx *gin.Context) {
	log.C(ctx).Infow("DeleteImage")

	imageUUID := ctx.Param("image_uuid")

	if imageUUID == "" || !govalidator.IsUUID(imageUUID) {
		core.WriteResponse(ctx, errno.ErrInvalidParameter, nil)
		return
	}

	jwtUserUUID, ok := controller.GetUserUUIDFromContext(ctx)
	if !ok {
		core.WriteResponse(ctx, errno.ErrTokenInvalid, nil)
		return
	}
	if err := ctrl.b.Images().Delete(ctx, jwtUserUUID, imageUUID); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, nil)
}
