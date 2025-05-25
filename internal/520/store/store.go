package store

import (
	"gorm.io/gorm"
	"sync"
)

var (
	once sync.Once
	S    *datastore
)

type IStore interface {
	DB() *gorm.DB
	User() UserStore
	Image() ImageStore
}

type datastore struct {
	db *gorm.DB
}

var _ IStore = (*datastore)(nil)

func NewStore(db *gorm.DB) IStore {
	once.Do(func() {
		S = &datastore{db: db}
	})
	return S
}

func (s *datastore) DB() *gorm.DB {
	return s.db
}

func (s *datastore) User() UserStore {
	return newUserStore(s.db)
}

func (s *datastore) Image() ImageStore {
	return newImageStore(s.db)
}
