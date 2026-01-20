package consts

import (
	"encoding/json"
	"log"

	"github.com/AnimateAIPlatform/animate-ai/common/env"
)

type EnvsConfig struct {
	ServerPort      string
	PrometheusPort  string
	ApolloAppID     string
	ApolloCluster   string
	ApolloNamespace string
	ApolloMetaAddr  string
}

var GlobalEnvs *EnvsConfig
var envConfFile = EnvConfFile

// SetEnvConfFile 设置环境配置文件路径
func SetEnvConfFile(confFile string) {
	envConfFile = confFile
}

func Init() error {
	// 加载 .env 文件，但不强制要求任何环境变量
	// 使用 LoadEnv 加载文件，如果文件不存在也不报错
	if envConfFile != "" {
		// 尝试加载 .env 文件，如果文件不存在或读取失败，忽略错误继续使用环境变量默认值
		_ = env.LoadEnv(envConfFile, []string{})
	}

	// 所有配置都是可选的，使用默认值或空字符串
	GlobalEnvs = &EnvsConfig{
		ServerPort:      env.GetEnvWithDefault("ServerPort", "10000"),
		PrometheusPort:  env.GetEnvWithDefault("PrometheusPort", "10001"),
		ApolloAppID:     env.GetEnvWithDefault("ApolloAppID", ""),
		ApolloCluster:   env.GetEnvWithDefault("ApolloCluster", ""),
		ApolloNamespace: env.GetEnvWithDefault("ApolloNamespace", ""),
		ApolloMetaAddr:  env.GetEnvWithDefault("ApolloMetaAddr", ""),
	}

	envJSON, err := json.Marshal(GlobalEnvs)
	if err != nil {
		return err
	}
	log.Printf("env config: %s\n", string(envJSON))
	return nil
}
