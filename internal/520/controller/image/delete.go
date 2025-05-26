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

	userUUID, ok := ctx.Value("useruuid").(string)
	if !ok {
		core.WriteResponse(ctx, errno.ErrUnauthorized, nil)
		return
	}
	if err := ctrl.b.Images().Delete(ctx, userUUID, ctx.Param("imageId")); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, nil)
}
