package batchsaver

import "time"

// Config 批量处理器配置
type Config struct {
	DB            interface{}   // *gorm.DB
	TableName     string        // 表名（可选）
	UniqueKeys    []string      // 唯一约束字段
	BatchSize     int           // 批次大小
	FlushInterval time.Duration // 刷新间隔
}

// IBatchSaver 公共接口（用于管理器存储）
type IBatchSaver interface {
	Close() error
	Flush() error
}
