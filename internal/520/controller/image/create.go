package image

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/pkg/api"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) Create(ctx *gin.Context) {
	log.C(ctx).Infow("Create Image")

	var req api.CreateImageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(ctx, errno.ErrBind, nil)
		return
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(ctx, errno.ErrInvalidParameter, nil)
		return
	}

	respo, err := ctrl.b.Images().Create(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, respo)
}
