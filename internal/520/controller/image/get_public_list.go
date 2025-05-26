package image

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/log"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"strconv"
)

func (ctrl *ImageController) GetPublicList(ctx *gin.Context) {
	log.C(ctx).Infow("Get Public List")
	limit := ctx.DefaultQuery("limit", "10")
	if limit == "" && !govalidator.IsNumeric(limit) {
		limit = "10"
	}
	limitInt, err := strconv.Atoi(limit)
	resp, err := ctrl.b.Images().ListRandomPublicImages(ctx, limitInt)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, resp)
}
