package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ImageTagM struct {
	ID        uint              `gorm:"primary_key"`
	Tag       string            `gorm:"type:varchar(255);column:tag;not null;index:tag_image;collate:utf8mb4_unicode_ci" json:"tag"`
	ImageUUID datatypes.BinUUID `gorm:"type:binary(16);index:tag_image;column:imageUUID"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *ImageTagM) TableName() string {
	return "image_tags"
}

func (u *ImageTagM) ToString() string {
	if u == nil {
		return ""
	}
	return u.Tag
}
