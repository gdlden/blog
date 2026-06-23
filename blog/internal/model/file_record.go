package model

import "gorm.io/gorm"

type FileRecord struct {
	gorm.Model
	FileName    string `gorm:"column:file_name;type:varchar(255);not null;comment:原始文件名"`
	FilePath    string `gorm:"column:file_path;type:varchar(512);not null;comment:存储路径/对象键"`
	FileType    string `gorm:"column:file_type;type:varchar(128);comment:MIME类型"`
	FileExt     string `gorm:"column:file_ext;type:varchar(32);comment:文件后缀"`
	FileUrl     string `gorm:"column:file_url;type:text;comment:访问URL"`
	FileSize    int64  `gorm:"column:file_size;default:0;comment:文件大小(字节)"`
	StorageType string `gorm:"column:storage_type;type:varchar(32);comment:存储后端类型"`
}

func (FileRecord) TableName() string {
	return "file_records"
}
