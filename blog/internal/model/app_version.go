package model

import "gorm.io/gorm"

type AppVersion struct {
	gorm.Model
	Version    string `gorm:"column:version;type:varchar(32);not null;index;comment:版本号"`
	Info       string `gorm:"column:info;type:text;comment:更新说明(JSON数组)"`
	IosUrl     string `gorm:"column:ios_url;type:text;comment:iOS下载地址"`
	AndroidUrl string `gorm:"column:android_url;type:text;comment:Android下载地址"`
	IsActive   bool   `gorm:"column:is_active;default:false;comment:是否当前激活版本"`
}

func (AppVersion) TableName() string {
	return "app_versions"
}
