package apollo

import (
	"sync"

	"github.com/philchia/agollo/v4"
)

// 监听器类型，接收 namespace, key 和 change
type ConfigChangeListener func(namespace, key string, change *agollo.Change)

// 全局监听器
type ListenerManager struct {
	listenersMu sync.RWMutex
	listeners   []ConfigChangeListener
}

var globalListenerManager = &ListenerManager{}

// 注册监听器
func RegisterListener(listener ConfigChangeListener) {
	globalListenerManager.listenersMu.Lock()
	defer globalListenerManager.listenersMu.Unlock()
	globalListenerManager.listeners = append(globalListenerManager.listeners, listener)
}

// 通知监听器
func notifyListeners(namespace, key string, change *agollo.Change) {
	globalListenerManager.listenersMu.RLock()
	defer globalListenerManager.listenersMu.RUnlock()
	for _, listener := range globalListenerManager.listeners {
		// 调用每个监听器
		listener(namespace, key, change)
	}
}
