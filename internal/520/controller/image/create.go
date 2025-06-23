package image

import (
	"demo520/internal/520/controller"
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/pkg/api"

	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) Create(ctx *gin.Context) {
	defer log.FuncEntryWithContext(ctx.Request.Context())()

	// 从JWT中获取用户UUID
	userUUID, exists := controller.GetUserUUIDFromContext(ctx)
	if !exists {
		err := errno.ErrTokenInvalid
		log.ErrorWithFunc(err, "无法从JWT中获取用户UUID")
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 解析表单数据
	var r api.CreateImageRequest
	if err := ctx.ShouldBind(&r); err != nil {
		log.ErrorWithFunc(err, "请求参数绑定失败")
		core.WriteResponse(ctx, errno.ErrBind, nil)
		return
	}

	// 获取上传的文件
	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		err := errno.ErrInvalidParameter.SetMessage("image file is required")
		log.ErrorWithFunc(err, "获取上传文件失败")
		core.WriteResponse(ctx, err, nil)
		return
	}
	// 调用业务层创建图片
	resp, err := ctrl.b.Images().Create(ctx.Request.Context(), userUUID, &r, fileHeader)
	if err != nil {
		log.ErrorWithFunc(err, "创建图片失败", "userUUID", userUUID)
		core.WriteResponse(ctx, err, nil)
		return
	}

	log.Infow("图片创建成功", "userUUID", userUUID, "imageUUID", resp.ImageUUID)
	core.WriteResponse(ctx, nil, resp)
}
