package dao

import (
	"fmt"

	"github.com/AnimateAIPlatform/animate-ai/common/apollo"
	"github.com/AnimateAIPlatform/animate-ai/models"
)

func GetStaticDBConfigApp(app string) (*models.AppClusterInfo, error) {
	appClusterInfo := &models.StaticAppClusterInfo{}
	err := apollo.GetValueFromEnvAndApollo(appClusterInfo)
	if err != nil {
		return nil, fmt.Errorf("获取应用环境配置失败: %w", err)
	}

	appEnvConfig := models.AppClusterInfo{}
	for _, info := range appClusterInfo.Data {
		if info.App == app {
			appEnvConfig = info
			return &appEnvConfig, nil
		}
	}

	return nil, fmt.Errorf("没有这个app: %+v", app)

}
