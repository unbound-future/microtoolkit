package models

import (
	"gorm.io/gorm"
)

// WorkflowTemplate 工作流模版表
type WorkflowTemplate struct {
	gorm.Model
	UserID    string `gorm:"type:varchar(100);not null;index" json:"user_id"`        // 用户ID
	TemplateID string `gorm:"type:varchar(100);not null;uniqueIndex" json:"template_id"`  // 模版ID（唯一）
	Name      string `gorm:"type:varchar(255);not null" json:"name"`                 // 模版名称
	Description string `gorm:"type:text" json:"description,omitempty"`             // 模版描述（可选）
	AssetID   string `gorm:"type:varchar(100);index" json:"asset_id,omitempty"`      // 关联的资产ID（可选）
	TemplateData  string `gorm:"type:longtext;not null" json:"template_data"`                // 模版数据（JSON格式）
}

// TableName 指定表名
func (WorkflowTemplate) TableName() string {
	return "workflow_templates"
}
