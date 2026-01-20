package ruleengine

import (
	"regexp"
	"sync/atomic"

	"github.com/AnimateAIPlatform/animate-ai/common/apollo"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/philchia/agollo/v4"
)

type Reg struct {
	Object            string `json:"object"`
	RegularExpression string `json:"regular-expression"`
}

type ConfigHolder[T any] struct {
	key   string
	value atomic.Value
}

func NewConfigHolder[T any](apolloKey string) *ConfigHolder[T] {
	return &ConfigHolder[T]{key: apolloKey}
}

func (c *ConfigHolder[T]) Init() error {
	var cfg T
	// 优先从本地配置文件读取，如果本地文件不存在则从Apollo读取
	if err := apollo.LoadConfigWithLocalFirst(apollo.ApolloNamespaceApplication, map[string]interface{}{c.key: &cfg}); err != nil {
		return err
	}
	c.value.Store(cfg)
	hlog.Infof("Initial config loaded for %s: %+v", c.key, cfg)

	// 注册Apollo配置变更监听器（仅当配置来自Apollo时生效）
	apollo.RegisterListener(func(namespace, key string, change *agollo.Change) {
		if key == c.key {
			// 更新配置（仅从Apollo更新，本地文件不支持热更新）
			hlog.Infof("%s updated, updating config", c.key)
			var newCfg T
			apollo.UpdateConfigWithNamespace(apollo.ApolloNamespaceApplication, map[string]interface{}{c.key: &newCfg})
			c.value.Store(newCfg)
			hlog.Infof("Config updated for %s: %+v", c.key, newCfg)
		}
	})

	return nil
}

func (c *ConfigHolder[T]) Get() T {
	cfg := c.value.Load()
	if cfg == nil {
		var zero T
		return zero
	}
	return cfg.(T)
}

// MatchRegs 判断 content 是否同时匹配所有正则
func MatchRegs(regs []Reg, content map[string]string) bool {
	for _, reg := range regs {
		matched, err := regexp.MatchString(reg.RegularExpression, content[reg.Object])
		if err != nil || !matched {
			return false
		}
	}
	return true
}
