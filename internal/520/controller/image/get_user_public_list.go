package image

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) GetUserPublicList(ctx *gin.Context) {
	log.C(ctx).Infow("Get Public List")

	var listRange ListRange
	if err := ctx.ShouldBindQuery(&listRange); err != nil {
		core.WriteResponse(ctx, errno.ErrBind, nil)
		return
	}

	// 获取公开图片，无需认证。直接从参数中解析UUID即可
	var userUUID string
	if userUUID = ctx.Param("userUUID"); userUUID == "" || !govalidator.IsUUID(userUUID) {
		core.WriteResponse(ctx, errno.ErrBind, nil)
		return
	}

	resp, err := ctrl.b.Images().ListUserOwnPublicImages(ctx, userUUID, listRange.Offset, listRange.Limit)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, resp)
}
