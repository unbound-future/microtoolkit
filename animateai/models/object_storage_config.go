package models

import (
	"gorm.io/gorm"
)

// ObjectStorageConfig 对象存储配置表
type ObjectStorageConfig struct {
	gorm.Model
	Name        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"` // 配置名称（唯一）
	Type        string `gorm:"type:varchar(50);not null" json:"type"`              // 存储类型：oss, s3, cos, minio 等
	Endpoint    string `gorm:"type:varchar(255);not null" json:"endpoint"`         // 存储服务端点
	Bucket      string `gorm:"type:varchar(100);not null" json:"bucket"`           // 存储桶名称
	AccessKey   string `gorm:"type:varchar(255);not null" json:"access_key"`       // 访问密钥ID
	SecretKey   string `gorm:"type:varchar(255);not null" json:"secret_key"`       // 访问密钥
	Region      string `gorm:"type:varchar(50)" json:"region,omitempty"`           // 区域（可选）
	BaseURL     string `gorm:"type:varchar(500)" json:"base_url,omitempty"`        // 基础URL（可选，用于自定义域名）
	IsDefault   bool   `gorm:"default:false;index" json:"is_default"`              // 是否为默认配置
	Status      int    `gorm:"default:1;index" json:"status"`                      // 状态：1=启用, 0=禁用
	Description string `gorm:"type:varchar(500)" json:"description,omitempty"`     // 描述信息
}

// TableName 指定表名
func (ObjectStorageConfig) TableName() string {
	return "object_storage_configs"
}
