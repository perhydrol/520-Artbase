package image

import (
	"demo520/internal/520/controller"
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/pkg/api"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) UpdateImageTags(ctx *gin.Context) {
	log.C(ctx).Infow("UpdateImageTags")
	var req api.UpdateImageTagsRequest

	imageUUID := ctx.Param("image_uuid")
	if err := ctx.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(ctx, errno.ErrBind, nil)
		return
	}

	if _, err := govalidator.ValidateStruct(req); err != nil || !govalidator.IsUUIDv4(imageUUID) {
		core.WriteResponse(ctx, errno.ErrInvalidParameter, nil)
		return
	}

	userUUID, exists := controller.GetUserUUIDFromContext(ctx)
	if !exists {
		err := errno.ErrTokenInvalid
		log.ErrorWithFunc(err, "无法从JWT中获取用户UUID")
		core.WriteResponse(ctx, err, nil)
		return
	}

	if err := ctrl.b.Images().UpdateTags(ctx, userUUID, imageUUID, &req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, nil)
}
