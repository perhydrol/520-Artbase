package image

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) Get(ctx *gin.Context) {
	log.C(ctx).Infow("Get image")
	var imageUUID string = ctx.Param("imageuuid")

	if govalidator.IsUUIDv4(imageUUID) == false {
		core.WriteResponse(ctx, errno.ErrInvalidParameter, nil)
		return
	}

	resp, err := ctrl.b.Images().Get(ctx, imageUUID)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, resp)
}
