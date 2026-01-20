package models

import (
	"encoding/json"
	"fmt"
)

type MySQLDBConfig string

type DBConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
}

type StaticNewapiDBKey struct {
	DBConfig
}

func (s *StaticNewapiDBKey) GetPortString() string {
	return fmt.Sprintf("%d", s.Port)
}

func (s *StaticNewapiDBKey) GetKey() string {
	return "static_newapi_db"
}

func (s *StaticNewapiDBKey) GetNamespace() string {
	return "application"
}

func (s *StaticNewapiDBKey) GetEnvOverrideKey() string {
	return "STATIC_NEWAPI_DB_OVERRIDE"
}

func (s *StaticNewapiDBKey) UnmarshalToValue(data string) error {
	return json.Unmarshal([]byte(data), s)
}

type StaticDBConfigKey struct {
	DBConfig
}

// GetPortString 获取端口号的字符串表示
func (d *StaticDBConfigKey) GetPortString() string {
	return fmt.Sprintf("%d", d.Port)
}

func (s *StaticDBConfigKey) GetKey() string {
	return "static_db_config"
}

func (s *StaticDBConfigKey) GetNamespace() string {
	return "application"
}

func (s *StaticDBConfigKey) GetEnvOverrideKey() string {
	return "STATIC_DB_CONFIG"
}

func (s *StaticDBConfigKey) UnmarshalToValue(data string) error {
	return json.Unmarshal([]byte(data), s)
}

type StaticAppClusterInfo struct {
	Data []AppClusterInfo `json:"data"`
}

type AppClusterInfo struct {
	App  string `json:"app"`
	Host string `json:"host"`
	Port string `json:"port"`
}

func (s *StaticAppClusterInfo) GetKey() string {
	return "static_app_cluster_info"
}

func (s *StaticAppClusterInfo) GetNamespace() string {
	return "RD-1.env_base"
}

func (s *StaticAppClusterInfo) GetEnvOverrideKey() string {
	return "APP_CLUSTER_CONFIG_OVERRIDE"
}

func (s *StaticAppClusterInfo) UnmarshalToValue(data string) error {
	return json.Unmarshal([]byte(data), &s.Data)
}
