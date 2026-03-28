package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID          string `gorm:"primarykey"`
	Username    string `gorm:"column:username"`
	Password    string `gorm:"column:password"`
	PhoneNumber string `gorm:"column:phone_number"`
	Email       string `gorm:"column:email"`
	Avatar      string `gorm:"column:avatar"`
}

func (User) TableName() string {
	return "users"
}
