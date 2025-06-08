package model

import (
	"cmp"
	"gorm.io/gorm"
	"slices"
	"time"
)

type ImageM struct {
	ImageUUID string      `gorm:"type:char(36);column:imageUUID;primaryKey" json:"imageuuid"`
	Hash      string      `gorm:"type:char(64);column:hash;index;not null" json:"hash"`
	Token     string      `gorm:"type:char(36);column:token;uniqueIndex" json:"token"`
	UserUUID  string      `gorm:"type:char(36);column:userUUID;not null" json:"useruuid"`
	IsPublic  bool        `gorm:"type:boolean;column:is_public;not null" json:"is_public"`
	Tags      []ImageTagM `gorm:"foreignKey:ImageUUID;references:ImageUUID" json:"tags"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *ImageM) TableName() string {
	return "images"
}

func (u *ImageM) Equal(other *ImageM) bool {
	if u == nil || other == nil {
		return false
	}
	if u.ImageUUID != other.ImageUUID &&
		u.Hash != other.Hash &&
		u.Token != other.Token &&
		u.UserUUID != other.UserUUID &&
		u.IsPublic != other.IsPublic &&
		len(u.Tags) != len(other.Tags) {
		return false
	}

	slices.SortFunc(u.Tags, func(a, b ImageTagM) int { return cmp.Compare(a.Tag, b.Tag) })
	slices.SortFunc(other.Tags, func(a, b ImageTagM) int { return cmp.Compare(a.Tag, b.Tag) })

	return slices.EqualFunc(u.Tags, other.Tags, func(a, b ImageTagM) bool { return a.Tag == b.Tag })
}
