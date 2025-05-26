package user

import (
	"demo520/internal/520/biz"
	"demo520/internal/520/store"
)

type UserController struct {
	b biz.IBiz
}

func NewUserController(db store.IStore) *UserController {
	return &UserController{biz.NewIBiz(db)}
}
