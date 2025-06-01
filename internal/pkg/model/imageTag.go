package model

import (
	"gorm.io/gorm"
	"time"
)

type ImageTagM struct {
	ID        uint   `gorm:"primary_key"`
	Tag       string `gorm:"type:varchar(255);column:tag;not null;index:tag_image;collate:utf8mb4_unicode_ci" json:"tag"`
	ImageUUID string `gorm:"type:char(36);not null;index:tag_image"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *ImageTagM) TableName() string {
	return "imageTag"
}
