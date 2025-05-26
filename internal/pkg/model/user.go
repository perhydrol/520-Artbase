package model

import (
	"demo520/internal/pkg/log"
	"demo520/pkg/auth"
	"gorm.io/gorm"
	"time"
)

type UserM struct {
	UserUUID  string    `gorm:"type:char(36);column:userUUID;not null;<-:create;primary_key" json:"useruuid"`
	Password  string    `gorm:"type:char(32);column:password;not null" json:"-"`
	Nickname  string    `gorm:"type:varchar(100);column:nickname;collate:utf8mb4_unicode_ci" json:"nickname"`
	Email     string    `gorm:"type:varchar(255);column:email;unique;index" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u *UserM) TableName() string {
	return "users"
}

func (u *UserM) BeforeCreate(tx *gorm.DB) error {
	err := error(nil)
	u.Password, err = auth.HashPassword(u.Password)
	if err != nil {
		log.Errorw("Error hashing password")
		return err
	}
	return nil
}
