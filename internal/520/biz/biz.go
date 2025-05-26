package biz

import (
	"demo520/internal/520/biz/image"
	"demo520/internal/520/biz/user"
	"demo520/internal/520/store"
)

type IBiz interface {
	Images() image.ImageBiz
	Users() user.UserBiz
}

type biz struct {
	db store.IStore
}

var _ IBiz = (*biz)(nil)

func NewIBiz(db store.IStore) IBiz {
	return &biz{db}
}

func (b *biz) Images() image.ImageBiz {
	return nil
}

func (b *biz) Users() user.UserBiz {
	return nil
}
