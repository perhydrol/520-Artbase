package image

import (
	"demo520/internal/pkg/core"
	"demo520/internal/pkg/errno"
	"demo520/internal/pkg/log"
	"demo520/pkg/api"
	"encoding/json"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func (ctrl *ImageController) Create(ctx *gin.Context) {
	log.C(ctx).Infow("Create Image")

	var req api.CreateImageRequest
	metadataStr := ctx.PostForm("json")
	if metadataStr == "" {
		core.WriteResponse(ctx, errno.ErrImageJSONNotFound, nil)
		return
	}
	if !govalidator.IsJSON(metadataStr) {
		core.WriteResponse(ctx, errno.ErrImageJSONInvalid, nil)
		return
	}

	if err := json.Unmarshal([]byte(metadataStr), &req); err != nil {
		core.WriteResponse(ctx, errno.ErrBind, nil)
		return
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		core.WriteResponse(ctx, errno.ErrInvalidParameter, nil)
		return
	}

	jwtUserUUID, err := pareseJwtAndEqualReqUUID(ctx, &req)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	respo, err := ctrl.b.Images().Create(ctx, jwtUserUUID, &req, file)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, respo)
}
