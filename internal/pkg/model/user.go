package model

import (
	"demo520/internal/pkg/log"
	"demo520/pkg/auth"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint      `gorm:"column:id;primary_key" json:"id"`
	UserUUID  string    `gorm:"column:username;not null;<-:create" json:"useruuid"`
	Password  string    `gorm:"column:password;not null" json:"password"`
	Nickname  string    `gorm:"column:nickname" json:"nickname"`
	Email     string    `gorm:"column:email" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	err := error(nil)
	u.Password, err = auth.HashPassword(u.Password)
	if err != nil {
		log.Errorw("Error hashing password")
		return err
	}
	return nil
}
