package dao

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbInstances = make(map[string]*gorm.DB)
	mu          sync.Mutex
)

// GetDatabase 根据 dsn 获取 DB，支持多个实例+多个库
func GetDatabase(username, password, host, port, database string) (*gorm.DB, error) {
	// DSN 本身就唯一标识了实例+库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", username, password, host, port, database)

	mu.Lock()
	defer mu.Unlock()

	if db, ok := dbInstances[dsn]; ok {
		return db, nil
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // 输出到标准输出
			logger.Config{
				SlowThreshold: time.Second, // 慢查询阈值
				LogLevel:      logger.Warn, // 打印部分 SQL
				Colorful:      true,        // 彩色打印
			},
		),
	})
	if err != nil {
		return nil, err
	}

	dbInstances[dsn] = db
	return db, nil
}
