package client

import (
	"sync/atomic"
	"time"

	"crypto/tls"

	"github.com/AnimateAIPlatform/animate-ai/common/apollo"
	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/common/types"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/philchia/agollo/v4"
)

type HttpClientConfig struct {
	DialTimeout         types.Duration `json:"DialTimeout"`         //TCP 连接建立时的最大等待时间
	MaxConnsPerHost     int            `json:"MaxConnsPerHost"`     //每个主机允许的最大连接数
	MaxIdleConnDuration types.Duration `json:"MaxIdleConnDuration"` //空闲连接最大存活时间
	MaxConnDuration     types.Duration `json:"MaxConnDuration"`     //最大连接存活时间
	MaxConnWaitTimeout  types.Duration `json:"MaxConnWaitTimeout"`  //等待空闲连接的最大时间
	KeepAlive           bool           `json:"KeepAlive"`           //是否使用长连接
	ClientReadTimeout   types.Duration `json:"ClientReadTimeout"`   //读取响应的最大时间
	ResponseBodyStream  bool           `json:"ResponseBodyStream"`  //是否流式读取 body
	WriteTimeout        types.Duration `json:"WriteTimeout"`        //写入请求超时时间
}

var (
	HttpClientConfigData = HttpClientConfig{
		DialTimeout:         types.Duration{Duration: 5 * time.Second},
		MaxConnsPerHost:     512,
		MaxIdleConnDuration: types.Duration{Duration: 1 * time.Minute},
		MaxConnDuration:     types.Duration{Duration: 0}, // 0 表示不限制连接持续时间
		MaxConnWaitTimeout:  types.Duration{Duration: 0}, // 0 表示不限制等待空闲连接的时间
		KeepAlive:           true,
		ClientReadTimeout:   types.Duration{Duration: 0}, // 0 表示不限制读取响应的时间
		ResponseBodyStream:  true,
		WriteTimeout:        types.Duration{Duration: 0}, // 0 表示不限制写入请求的时间
	}
)

// activeClient 存储当前活跃的 *http.Client
var activeClient atomic.Value

// 初始化
func InitHttpClient() error {
	// 优先从本地配置文件读取，如果本地文件不存在则从Apollo读取
	err := apollo.LoadConfigWithLocalFirst(apollo.ApolloNamespaceApplication, map[string]interface{}{
		consts.HttpClientConfigKey: &HttpClientConfigData,
	})
	if err != nil {
		return err
	}
	// 创建初始的 http.Client
	client, err := newHttpClientFromConfig()
	if err != nil {
		return err
	}
	activeClient.Store(client)
	// 注册Apollo配置变更监听器（仅当配置来自Apollo时生效）
	apollo.RegisterListener(func(namespace, key string, change *agollo.Change) {
		// 注册配置变更监听器

		if key == consts.HttpClientConfigKey {
			// 更新动态常量（仅从Apollo更新，本地文件不支持热更新）
			hlog.Infof("%s updated in namespace %s, updating HTTP client configuration", consts.HttpClientConfigKey, namespace)
			apollo.UpdateConfigWithNamespace(apollo.ApolloNamespaceApplication, map[string]interface{}{
				consts.HttpClientConfigKey: &HttpClientConfigData,
			})
			// 新建 client 替换原有 client
			hlog.Infof("%s updated in namespace %s, creating new HTTP client", consts.HttpClientConfigKey, namespace)
			UpdateHttpClientOnConfigChange()
		}

	})

	return nil
}

// 获取当前 client
func GetClient() *client.Client {
	return activeClient.Load().(*client.Client)
}

// 根据当前配置创建新的 http.Client
func newHttpClientFromConfig() (*client.Client, error) {
	return client.NewClient(
		client.WithDialTimeout(HttpClientConfigData.DialTimeout.Duration),
		client.WithMaxConnsPerHost(HttpClientConfigData.MaxConnsPerHost),
		client.WithMaxIdleConnDuration(HttpClientConfigData.MaxIdleConnDuration.Duration),
		client.WithMaxConnDuration(HttpClientConfigData.MaxConnDuration.Duration),
		client.WithMaxConnWaitTimeout(HttpClientConfigData.MaxConnWaitTimeout.Duration),
		client.WithKeepAlive(HttpClientConfigData.KeepAlive),
		client.WithClientReadTimeout(HttpClientConfigData.ClientReadTimeout.Duration),
		client.WithResponseBodyStream(HttpClientConfigData.ResponseBodyStream),
		client.WithWriteTimeout(HttpClientConfigData.WriteTimeout.Duration),
		client.WithTLSConfig(&tls.Config{
			RootCAs: nil, // 使用系统 CA
		}),
	)
}

// Apollo 配置变更事件触发
func UpdateHttpClientOnConfigChange() {
	newClient, err := newHttpClientFromConfig()
	if err != nil {
		hlog.Errorf("Error creating new HTTP client: %v", err)
		return
	}
	activeClient.Store(newClient)
	hlog.Infof("HTTP client configuration updated, new client created with config: %+v",
		HttpClientConfigData)
	// 旧 client 的连接池继续服务在飞请求，GC 回收旧 client
}
