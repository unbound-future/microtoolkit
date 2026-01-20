package models

import (
	"gorm.io/gorm"
)

// ToolComponentType 工具组件类型
const (
	ToolComponentTypeAsset   = "asset"   // 资产组件
	ToolComponentTypeService = "service" // 服务组件
	ToolComponentTypeTrigger = "trigger" // 时间触发器组件
)

// ToolComponent 工具组件表
type ToolComponent struct {
	gorm.Model
	UserID      string `gorm:"type:varchar(100);not null;index" json:"user_id"`        // 用户ID
	ComponentID string `gorm:"type:varchar(100);not null;uniqueIndex" json:"component_id"` // 组件ID（唯一）
	Name        string `gorm:"type:varchar(255);not null" json:"name"`                 // 组件名称
	Description string `gorm:"type:text" json:"description,omitempty"`                 // 组件描述，可选
	Type        string `gorm:"type:varchar(20);not null;index" json:"type"`            // 组件类型：asset 或 service
	
	// 资产组件相关字段
	AssetID *string `gorm:"type:varchar(100);index" json:"asset_id,omitempty"` // 关联的资产ID（资产组件类型时使用）
	
	// 服务组件相关字段
	ServiceURL *string `gorm:"type:text" json:"service_url,omitempty"`       // 服务URL（服务组件类型时使用）
	ParamDesc  *string `gorm:"type:text" json:"param_desc,omitempty"`        // 参数说明（服务组件类型时使用）
	
	// 时间触发器组件相关字段
	CronExpression *string `gorm:"type:varchar(255)" json:"cron_expression,omitempty"` // Cron表达式（时间触发器类型时使用）
}

// TableName 指定表名
func (ToolComponent) TableName() string {
	return "tool_components"
}
