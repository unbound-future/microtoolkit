package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     int
	DBName   string
}

var (
	DB  *gorm.DB
	cfg *Config
)

func InitDB(c *Config) error {
	var err error
	cfg = c
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return err
}
