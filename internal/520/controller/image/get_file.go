package image

import (
	"demo520/internal/520/controller"
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"path/filepath"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) GetImageFile(ctx *gin.Context) {
	log.C(ctx).Infow("Get image file")
	var imageUUIDFileName string = ctx.Param("imageuuidFile")

	ext := filepath.Ext(imageUUIDFileName)
	imageUUID := strings.TrimSuffix(imageUUIDFileName, ext)
	if ext == "" || len(ext) >= len(imageUUIDFileName) {
		core.WriteResponse(ctx, errno.ErrImageFileInvalid, nil)
		return
	}
	if !govalidator.IsUUID(imageUUID) {
		core.WriteResponse(ctx, errno.ErrImageFileInvalid, nil)
		return
	}

	jwtUserUUID, ok := controller.GetUserUUIDFromContext(ctx)
	if !ok {
		jwtUserUUID = ""
	}
	hashFilePath, err := ctrl.b.Images().GetImageFile(ctx, jwtUserUUID, imageUUIDFileName)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.ResponseFile(ctx, nil, hashFilePath)
}
