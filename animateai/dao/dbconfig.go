package dao

import (
	"context"
	"encoding/json"

	"github.com/AnimateAIPlatform/animate-ai/common/apollo"
	"github.com/AnimateAIPlatform/animate-ai/models"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func GetStaticDBConfig(ctx context.Context, key models.MySQLDBConfig) *models.DBConfig {
	dbConfigJson, ok := apollo.GetGlobalConfigManager().GetVariableWithNamespace(apollo.ApolloNamespaceApplication, string(key))
	if !ok {
		hlog.CtxErrorf(ctx, "Error getting db config: %v", string(key)+" not found")
		return nil
	}

	// 打印原始配置JSON
	hlog.CtxInfof(ctx, "Raw db config JSON: %s", dbConfigJson)

	var dbConfig models.DBConfig
	err := json.Unmarshal([]byte(dbConfigJson), &dbConfig)
	if err != nil {
		hlog.CtxErrorf(ctx, "Error unmarshalling db config: %v", err)
		return nil
	}

	// 打印解析后的配置
	hlog.CtxInfof(ctx, "Parsed db config: Username=%s, Password=%s, Host=%s, Port=%d, Database=%s",
		dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)

	return &dbConfig
}
