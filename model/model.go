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
	Memorizes  []Memorize
}

type Memorize struct {
	gorm.Model
	UserID          uint
	SurahName       string
	AyahRange       string
	TotalAyah       int
	DateStarted     time.Time
	DateCompleted   time.Time
	ReviewFrequency string
	LastReviewDate  time.Time
	AccuracyLevel   string
	NextReviewDate  time.Time
	Notes           string
}
