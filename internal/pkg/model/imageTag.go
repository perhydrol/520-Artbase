package model

type ImageTag struct {
	ID        uint   `gorm:"primary_key"`
	Tag       string `gorm:"type:varchar(255);column:tag;not null;index:tag_image" json:"tag"`
	ImageUUID string `gorm:"type:char(36);not null;index:tag_image"`
	Image     Image  `gorm:"foreignKey:ImageUUID;references:ImageUUID" json:"image"`
}

func (u *ImageTag) TableName() string {
	return "imageTag"
}
