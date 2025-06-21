package image

import (
	"demo520/internal/520/controller"
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) Get(ctx *gin.Context) {
	log.C(ctx).Infow("Get image")
	var imageUUID string = ctx.Param("imageuuid")

	if !govalidator.IsUUIDv4(imageUUID) {
		core.WriteResponse(ctx, errno.ErrInvalidParameter, nil)
		return
	}

	jwtUserUUID, ok := controller.GetUserUUIDFromContext(ctx)
	if !ok {
		jwtUserUUID = ""
	}
	resp, err := ctrl.b.Images().Get(ctx, jwtUserUUID, imageUUID)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, resp)
}
