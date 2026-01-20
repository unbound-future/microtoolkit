package apollo

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// 通用解析 JSON 到结构体
func parseConfig(jsonStr string, target interface{}) error {
	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("parse config error: %v", err)
	}
	return nil
}

// loadConfigFromEnv 从环境变量读取JSON格式的配置
func loadConfigFromEnv(envKey string, target interface{}) error {
	configJson := os.Getenv(envKey)
	if configJson == "" {
		return fmt.Errorf("environment variable %s is not set", envKey)
	}

	if err := parseConfig(configJson, target); err != nil {
		return fmt.Errorf("error parsing %s from env: %v", envKey, err)
	}

	hlog.Infof("Loaded config from environment variable: %s", envKey)
	return nil
}

// LoadConfigWithLocalFirst 优先从环境变量读取，如果环境变量不存在或读取失败，则从Apollo读取
// namespace: Apollo命名空间
// configMap: 配置key和对应结构体指针的映射，key直接作为环境变量名（需要是大写格式）
func LoadConfigWithLocalFirst(namespace string, configMap map[string]interface{}) error {
	for key, target := range configMap {
		// 直接使用 key 作为环境变量名（应该是大写格式，如：DYNAMIC_HTTP_CLIENT_CONFIG）
		envKey := key

		// 首先尝试从环境变量读取（JSON格式）
		configJson := os.Getenv(envKey)
		if configJson != "" {
			err := loadConfigFromEnv(envKey, target)
			if err == nil {
				// 环境变量读取成功，使用环境变量配置
				hlog.Infof("%s loaded from environment variable %s: %+v", key, envKey, target)
				continue
			}
			// 环境变量存在但解析失败，记录警告并继续从Apollo读取
			hlog.Warnf("Failed to parse %s from environment variable %s: %v, trying Apollo", key, envKey, err)
		} else {
			// 环境变量不存在，这是正常情况，从Apollo读取
			hlog.Infof("%s not found in environment variable %s, loading from Apollo", key, envKey)
		}

		// 从Apollo读取
		configManager := GetGlobalConfigManager()
		if configManager == nil {
			return fmt.Errorf("apollo config manager is not initialized")
		}

		value, exists := configManager.GetVariableWithNamespace(namespace, key)
		if !exists {
			return fmt.Errorf("%s not found in Apollo config and environment variable", key)
		}

		if err := parseConfig(value, target); err != nil {
			return fmt.Errorf("error parsing %s: %v", key, err)
		}

		hlog.Infof("%s loaded from Apollo: %+v", key, target)
	}

	return nil
}

func UpdateConfigWithNamespace(namespace string, configMap map[string]interface{}) error {
	configManager := GetGlobalConfigManager()
	if configManager == nil {
		return fmt.Errorf("apollo config manager is not initialized")
	}

	for key, target := range configMap {
		value, exists := configManager.GetVariableWithNamespace(namespace, key)
		if !exists {
			return fmt.Errorf("%s not found in Apollo config", key)
		}

		if err := parseConfig(value, target); err != nil {
			return fmt.Errorf("error parsing %s: %v", key, err)
		}

		hlog.Infof("%s updated: %+v", key, target)
	}

	return nil
}
