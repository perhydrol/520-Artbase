package image

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) GetUserPublicList(ctx *gin.Context) {
	log.C(ctx).Infow("Get Public List")

	var listRange ListRange
	if err := ctx.ShouldBindQuery(&listRange); err != nil {
		core.WriteResponse(ctx, errno.ErrBind, nil)
		return
	}

	resp, err := ctrl.b.Images().ListUserOwnPublicImages(ctx, listRange.Offset, listRange.Limit)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, resp)
}
