package models

import (
	"gorm.io/gorm"
)

// AgentFlow 工作流表
type AgentFlow struct {
	gorm.Model
	UserID    string `gorm:"type:varchar(100);not null;index" json:"user_id"`        // 用户ID
	FlowID    string `gorm:"type:varchar(100);not null;uniqueIndex" json:"flow_id"`  // 工作流ID（唯一）
	Name      string `gorm:"type:varchar(255);not null" json:"name"`                 // 工作流名称
	AssetID   string `gorm:"type:varchar(100);index" json:"asset_id,omitempty"`      // 关联的资产ID（可选）
	TemplateID string `gorm:"type:varchar(100);index" json:"template_id,omitempty"`  // 工作流模版ID（可选）
	FlowData  string `gorm:"type:longtext;not null" json:"flow_data"`                // 工作流数据（JSON格式）
}

// TableName 指定表名
func (AgentFlow) TableName() string {
	return "agent_flows"
}



