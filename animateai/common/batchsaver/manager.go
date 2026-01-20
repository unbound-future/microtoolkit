package batchsaver

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"gorm.io/gorm"
)

// managerInstance 单例实例
var (
	managerInstance *BatchSaverManager
	managerOnce     sync.Once
	managerMu       sync.RWMutex
)

// BatchSaverManager 批量存储器管理器
type BatchSaverManager struct {
	savers sync.Map // key: typeString, value: IBatchSaver
}

// GetManager 获取管理器单例
func GetManager() *BatchSaverManager {
	managerOnce.Do(func() {
		managerInstance = &BatchSaverManager{}
	})
	return managerInstance
}

// GetOrCreateSaver 获取或创建批量存储器（包级泛型函数）
func GetOrCreateSaver[T any](
	db *gorm.DB,
	tableName string,
	uniqueKeys []string,
	batchSize int,
	flushInterval time.Duration,
) (*GenericBatchSaver[T], error) {

	// 获取类型字符串作为键
	var zero T
	typeKey := getTypeKey(zero)

	// 尝试从管理器中获取现有实例
	manager := GetManager()
	if saver, ok := manager.savers.Load(typeKey); ok {
		if typedSaver, ok := saver.(*GenericBatchSaver[T]); ok {
			return typedSaver, nil
		}
		// 类型不匹配，删除旧实例
		manager.savers.Delete(typeKey)
	}

	// 创建新实例
	saver, err := NewGenericBatchSaver[T](Config{
		DB:            db,
		TableName:     tableName,
		UniqueKeys:    uniqueKeys,
		BatchSize:     batchSize,
		FlushInterval: flushInterval,
	})
	if err != nil {
		return nil, fmt.Errorf("创建批量存储器失败: %w", err)
	}

	// 存储到管理器
	manager.savers.Store(typeKey, saver)
	return saver, nil
}

// getTypeKey 获取类型唯一键
func getTypeKey(v interface{}) string {
	t := reflect.TypeOf(v)
	// 使用包路径+类型名作为唯一键
	return t.PkgPath() + "." + t.Name()
}

// SaveModel 便捷函数：保存单个模型
func SaveModel[T any](
	db *gorm.DB,
	item T,
	tableName string,
	uniqueKeys []string,
) error {

	saver, err := GetOrCreateSaver[T](db, tableName, uniqueKeys, 1000, 5*time.Second)
	if err != nil {
		return err
	}
	return saver.Save(item)
}

// CloseAll 关闭所有存储器
func (m *BatchSaverManager) CloseAll() {
	m.savers.Range(func(key, value interface{}) bool {
		if saver, ok := value.(interface{ Close() error }); ok {
			saver.Close()
		}
		return true
	})
}
