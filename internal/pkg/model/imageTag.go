package model

type ImageTagM struct {
	ID        uint   `gorm:"primary_key"`
	Tag       string `gorm:"type:varchar(255);column:tag;not null;index:tag_image" json:"tag"`
	ImageUUID string `gorm:"type:char(36);not null;index:tag_image"`
	Image     ImageM `gorm:"foreignKey:ImageUUID;references:ImageUUID" json:"image"`
}

func (u *ImageTagM) TableName() string {
	return "imageTag"
}
