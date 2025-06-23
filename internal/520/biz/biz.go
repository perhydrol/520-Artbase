package biz

import (
	"demo520/internal/520/biz/image"
	"demo520/internal/520/biz/user"
	"demo520/internal/520/store"
	"sync"
)

var (
	once sync.Once
	ibiz *biz
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
	once.Do(func() {
		ibiz = &biz{db}
	})
	return ibiz
}

func (b *biz) Images() image.ImageBiz {
	return image.NewImageBiz(b.db)
}

func (b *biz) Users() user.UserBiz {
	return user.NewUserBiz(b.db)
}
