package models

import (
	"gorm.io/gorm"
)

// UserAsset 用户资产表
type UserAsset struct {
	gorm.Model
	UserID          string `gorm:"type:varchar(100);not null;index" json:"user_id"`        // 用户ID
	AssetID         string `gorm:"type:varchar(100);not null;uniqueIndex" json:"asset_id"` // 资产ID（唯一）
	Name            string `gorm:"type:varchar(255);not null" json:"name"`                 // 资产名称
	Description     string `gorm:"type:text" json:"description,omitempty"`                 // 资产描述，可选
	URL             string `gorm:"type:text;not null" json:"url"`                          // 资产URL
	Source          string `gorm:"type:varchar(20);not null;default:'url'" json:"source"`  // 资产来源：url 或 file
	Type            string `gorm:"type:varchar(20);not null;index" json:"type"`            // 资产类型：image, audio, video
	Size            *int64 `gorm:"type:bigint" json:"size,omitempty"`                      // 文件大小（字节），可选
	MimeType        string `gorm:"type:varchar(100)" json:"mime_type,omitempty"`           // MIME类型，可选
	StorageConfigID *uint  `gorm:"index" json:"storage_config_id,omitempty"`               // 对象存储配置ID（外键，可选）
	StorageURL      string `gorm:"type:text" json:"storage_url,omitempty"`                 // 对象存储URL（可选）
}

// TableName 指定表名
func (UserAsset) TableName() string {
	return "user_assets"
}
