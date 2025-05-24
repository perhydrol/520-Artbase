package model

import (
	"gorm.io/gorm"
	"time"
)

type ImageM struct {
	ImageUUID string      `gorm:"type:char(36);column:imageUUID;primaryKey" json:"imageuuid"`
	Hash      string      `gorm:"column:hash;index;not null" json:"hash"`
	Token     string      `gorm:"column:token;uniqueIndex" json:"token"`
	OwnerUUID string      `gorm:"type:char(36);column:ownerUUID;not null" json:"owneruuid"`
	IsPublic  bool        `gorm:"column:is_public;not null" json:"is_public"`
	Tags      []ImageTagM `gorm:"column:tags;foreignKey:ImageUUID;references:ImageUUID" json:"tags"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *ImageM) TableName() string {
	return "image"
}
