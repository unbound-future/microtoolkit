package cache

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/common/apollo"
	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/common/types"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/philchia/agollo/v4"
)

// Cache 内存缓存组件
type Cache struct {
	cache   *ristretto.Cache[string, any]
	baseTTL time.Duration
	jitter  time.Duration
}

// 配置
type CacheConfig struct {
	NumCounters int64          `json:"NumCounters"` // 统计计数器数量
	MaxCost     int64          `json:"MaxCost"`     // 最大缓存大小（字节）
	BufferItems int64          `json:"BufferItems"` // 内部队列缓冲
	BaseTTL     types.Duration `json:"BaseTTL"`     // 基础 TTL
	Jitter      types.Duration `json:"Jitter"`      // 随机浮动 TTL
}

var CacheConfigData = CacheConfig{
	NumCounters: 1e5,                                        // 统计计数器数量
	MaxCost:     1e7,                                        // 最大缓存大小（字节）(10 << 20约 10MB)
	BufferItems: 64,                                         // 内部队列缓冲
	BaseTTL:     types.Duration{Duration: 10 * time.Second}, // 基础 TTL
	Jitter:      types.Duration{Duration: 10 * time.Second}, // 随机浮动 TTL
}

var activeCache atomic.Value

func InitCache() error {
	// 优先从本地配置文件读取，如果本地文件不存在则从Apollo读取
	err := apollo.LoadConfigWithLocalFirst(apollo.ApolloNamespaceApplication, map[string]interface{}{
		consts.CacheConfigKey: &CacheConfigData,
	})
	if err != nil {
		return err
	}

	err = UpdateCache()
	if err != nil {
		return err
	}

	// 注册Apollo配置变更监听器（仅当配置来自Apollo时生效）
	apollo.RegisterListener(func(namespace, key string, change *agollo.Change) {
		if key == consts.CacheConfigKey {
			// 更新动态常量（仅从Apollo更新，本地文件不支持热更新）
			hlog.Infof("%s updated, updating cache config", consts.CacheConfigKey)
			apollo.UpdateConfigWithNamespace(apollo.ApolloNamespaceApplication, map[string]interface{}{
				consts.CacheConfigKey: &CacheConfigData,
			})
			// 重新创建缓存实例
			hlog.Infof("%s updated, creating new cache instance", consts.CacheConfigKey)
			err := UpdateCache()
			if err != nil {
				hlog.Errorf("failed to update cache: %v", err)
			} else {
				hlog.Infof("cache updated successfully")
			}
		}
	})

	return nil
}

// 初始化缓存组件
func UpdateCache() error {
	c, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: CacheConfigData.NumCounters,
		MaxCost:     CacheConfigData.MaxCost,
		BufferItems: CacheConfigData.BufferItems,
	})
	if err != nil {
		return err
	}

	activeCache.Store(&Cache{
		cache:   c,
		baseTTL: CacheConfigData.BaseTTL.Duration,
		jitter:  CacheConfigData.Jitter.Duration,
	})
	return nil
}

// 获取当前活跃的缓存实例
func GetCache() *Cache {
	return activeCache.Load().(*Cache)
}

// 生成随机 TTL（秒级粒度）
func (c *Cache) randomTTL() time.Duration {
	if c.jitter == 0 {
		return c.baseTTL
	}
	// 用秒级随机偏移，避免雪崩
	jitterSec := int64(c.jitter.Seconds())
	if jitterSec <= 0 {
		return c.baseTTL
	}
	return c.baseTTL + time.Duration(rand.Int63n(jitterSec))*time.Second
}

// 设置缓存
func (c *Cache) Get(key string) (any, bool) {
	if x, found := c.cache.Get(key); found {
		return x, true
	}
	return nil, false
}

// 获取缓存
func (c *Cache) Set(key string, value any) {
	c.cache.SetWithTTL(key, value, 1, c.randomTTL())
}

// 删除缓存
func (c *Cache) Delete(key string) {
	c.cache.Del(key)
}
