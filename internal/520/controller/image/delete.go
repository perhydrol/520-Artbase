package image

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) DeleteImage(ctx *gin.Context) {
	log.C(ctx).Infow("DeleteImage")

	var req string = ctx.Param("imageuuid")

	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(ctx, errno.ErrInvalidParameter, nil)
		return
	}

	jwtUserUUID, err := pareseJwtAndEqualReqUUID(ctx, nil)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	if err := ctrl.b.Images().Delete(ctx, jwtUserUUID, ctx.Param("imageId")); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, nil)
}
