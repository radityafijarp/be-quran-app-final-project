package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username   string `gorm:"uniqueIndex"`
	Password   string
	Fullname   string
	Desc       string
	ProfilePic string
	Photos     []Photo
}

type Photo struct {
	gorm.Model
	UserID    uint
	URL       string
	Caption   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
