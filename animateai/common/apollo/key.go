package apollo

import (
	"fmt"
	"os"
)

type EnvFirstInterfaceKey interface {
	GetKey() string
	GetNamespace() string

	GetEnvOverrideKey() string
	UnmarshalToValue(json string) error
}

func SetConfigFromEnv(envFirstInterfaceKey EnvFirstInterfaceKey) bool {
	// 从环境变量获取JSON格式的应用配置
	configJson := os.Getenv(envFirstInterfaceKey.GetEnvOverrideKey())
	if configJson == "" {
		return false
	}

	err := envFirstInterfaceKey.UnmarshalToValue(configJson)
	return err == nil
}

func GetValueFromEnvAndApollo(envFirstInterfaceKey EnvFirstInterfaceKey) error {
	// 优先从环境变量获取配置（用于拦截和覆盖）
	if found := SetConfigFromEnv(envFirstInterfaceKey); found {
		return nil
	}

	// 从Apollo配置中心获取配置
	appEnvConfigJson, ok := GetGlobalConfigManager().GetVariableWithNamespace(envFirstInterfaceKey.GetNamespace(), envFirstInterfaceKey.GetKey())
	if !ok {
		return fmt.Errorf("获取应用环境配置失败 %s %s", envFirstInterfaceKey.GetKey(), envFirstInterfaceKey.GetNamespace())
	}

	err := envFirstInterfaceKey.UnmarshalToValue(appEnvConfigJson)
	if err != nil {
		return err
	}
	return nil
}
