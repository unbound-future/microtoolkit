package models

import (
	"github.com/AnimateAIPlatform/animate-ai/common/db"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// InitTables 自动创建表
func InitTables() error {
	if db.DB == nil {
		hlog.Warnf("Database not initialized, skipping table creation")
		return nil
	}

	// 自动迁移表结构
	err := db.DB.AutoMigrate(
		&ObjectStorageConfig{},
		&UserAsset{},
		&User{},
		&ToolComponent{},
		&AgentFlow{},
		&WorkflowTemplate{},
	)
	if err != nil {
		return err
	}
	hlog.Infof("Tables auto-migrated successfully: object_storage_configs, user_assets, users, tool_components, agent_flows, workflow_templates")

	return nil
}
