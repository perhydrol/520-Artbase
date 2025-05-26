package image

import (
	"demo520/internal/520/biz"
	"demo520/internal/520/store"
)

type ImageController struct {
	b biz.IBiz
}

func NewUserController(db store.IStore) *ImageController {
	return &ImageController{biz.NewIBiz(db)}
}
