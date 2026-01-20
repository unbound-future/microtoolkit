package apollo

import (
	"log"
	"sync"

	"github.com/philchia/agollo/v4"
)

// ConfigUpdater 定义配置更新接口
type ConfigUpdater interface {

	// UpdateVariableWithNamespace 更新指定namespace的配置变量
	UpdateVariableWithNamespace(namespace, key string, value string, changeType agollo.ChangeType) error
	// GetVariable 获取指定配置变量的值

	GetVariableWithNamespace(namespace, key string) (string, bool)
	// GetAllVariables 获取所有配置变量
	GetAllVariables() map[string]string
	// GetVariablesByNamespace 获取指定namespace的所有配置变量
	GetVariablesByNamespace(namespace string) map[string]string
}

// ConfigManager 配置管理器实现
type ConfigManager struct {
	variables map[string]string // key格式: namespace.key 或 key (向后兼容)
	mutex     sync.RWMutex
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		variables: make(map[string]string),
		mutex:     sync.RWMutex{},
	}
}

// buildKey 构建带namespace前缀的key
func buildKey(namespace, key string) string {
	if namespace == "" {
		return key
	}
	return namespace + "." + key
}

// parseKey 解析key，返回namespace和原始key
func parseKey(fullKey string) (namespace, key string) {
	for i := len(fullKey) - 1; i >= 0; i-- {
		if fullKey[i] == '.' {
			return fullKey[:i], fullKey[i+1:]
		}
	}
	return "", fullKey
}

// UpdateVariableWithNamespace 更新指定namespace的配置变量
func (cm *ConfigManager) UpdateVariableWithNamespace(namespace, key string, value string, changeType agollo.ChangeType) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	fullKey := buildKey(namespace, key)

	// 更新内部变量存储
	switch changeType {
	case agollo.ADD, agollo.MODIFY:
		cm.variables[fullKey] = value
	case agollo.DELETE:
		delete(cm.variables, fullKey)
	}

	return nil
}

// GetVariableWithNamespace 获取指定namespace的配置变量
func (cm *ConfigManager) GetVariableWithNamespace(namespace, key string) (string, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	fullKey := buildKey(namespace, key)
	value, exists := cm.variables[fullKey]
	return value, exists
}

// GetAllVariables 获取所有配置变量
func (cm *ConfigManager) GetAllVariables() map[string]string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make(map[string]string)
	for k, v := range cm.variables {
		result[k] = v
	}
	return result
}

// GetVariablesByNamespace 获取指定namespace的所有配置变量
func (cm *ConfigManager) GetVariablesByNamespace(namespace string) map[string]string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make(map[string]string)
	prefix := namespace + "."

	for k, v := range cm.variables {
		if namespace == "" {
			// 如果没有指定namespace，返回所有不带namespace前缀的变量
			if _, key := parseKey(k); key == k {
				result[k] = v
			}
		} else {
			// 返回指定namespace的变量
			if len(k) > len(prefix) && k[:len(prefix)] == prefix {
				key := k[len(prefix):]
				result[key] = v
			}
		}
	}
	return result
}

// 全局配置管理器实例
var (
	globalConfigManager *ConfigManager
	globalOnce          sync.Once
)

// GetGlobalConfigManager 获取全局配置管理器
func GetGlobalConfigManager() *ConfigManager {
	globalOnce.Do(func() {
		globalConfigManager = NewConfigManager()
	})
	return globalConfigManager
}

func Init(appID, cluster string, namespaces []string, metaAddr string) (*agollo.Client, error) {
	apollo := agollo.NewClient(&agollo.Conf{
		AppID:          appID,
		Cluster:        cluster,
		NameSpaceNames: namespaces,
		MetaAddr:       metaAddr,
	})

	err := apollo.Start()
	if err != nil {
		return nil, err
	}

	for _, namespace := range namespaces {
		allKeys := apollo.GetAllKeys(agollo.WithNamespace(namespace))
		for _, key := range allKeys {
			value := apollo.GetString(key, agollo.WithNamespace(namespace))
			// 使用带namespace前缀的方式存储，避免不同namespace的相同key相互覆盖
			GetGlobalConfigManager().UpdateVariableWithNamespace(namespace, key, value, agollo.MODIFY)
		}
	}

	apollo.OnUpdate(func(event *agollo.ChangeEvent) {
		// 处理配置更新事件
		log.Printf("Apollo配置已更新 - Namespace: %s", event.Namespace)

		// 遍历所有变更的配置项
		for key, change := range event.Changes {
			switch change.ChangeType {
			case agollo.ADD:
				log.Printf("新增配置项: %s = %s", key, change.NewValue)
			case agollo.MODIFY:
				log.Printf("修改配置项: %s, 旧值: %s, 新值: %s", key, change.OldValue, change.NewValue)
			case agollo.DELETE:
				log.Printf("删除配置项: %s, 旧值: %s", key, change.OldValue)
			}

			// 使用全局配置管理器更新变量，包含namespace信息
			configManager := GetGlobalConfigManager()
			if err := configManager.UpdateVariableWithNamespace(event.Namespace, key, change.NewValue, change.ChangeType); err != nil {
				log.Printf("更新配置变量失败 - Namespace: %s, Key: %s, Error: %v", event.Namespace, key, err)
			}

			// 通知监听器
			go notifyListeners(event.Namespace, key, change)
		}
	})

	return &apollo, nil
}

// GetConfig 获取指定namespace的配置值
func GetConfig(namespace, key string) (string, bool) {
	return GetGlobalConfigManager().GetVariableWithNamespace(namespace, key)
}

// GetConfigWithDefault 获取指定namespace的配置值，如果不存在则返回默认值
func GetConfigWithDefault(namespace, key, defaultValue string) string {
	if value, exists := GetConfig(namespace, key); exists {
		return value
	}
	return defaultValue
}

// GetAllConfigsByNamespace 获取指定namespace的所有配置
func GetAllConfigsByNamespace(namespace string) map[string]string {
	return GetGlobalConfigManager().GetVariablesByNamespace(namespace)
}

// GetAllConfigs 获取所有配置（包含namespace前缀）
func GetAllConfigs() map[string]string {
	return GetGlobalConfigManager().GetAllVariables()
}
