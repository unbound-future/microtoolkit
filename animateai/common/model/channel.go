package model

import "encoding/json"

// Channel 对应 channels 表
type Channel struct {
    ID           int64           `gorm:"primaryKey;autoIncrement" json:"id"`
    Type         string          `gorm:"type:varchar(50);not null" json:"type"`
    ChannelKey   string          `gorm:"type:longtext;not null" json:"channel_key"`
    Status       int             `gorm:"default:1" json:"status"`
    Name         string          `gorm:"type:varchar(200);not null" json:"name"`
    CreatedTime  uint            `gorm:"not null" json:"created_time"`
    BaseURL      string          `gorm:"type:varchar(200);default:''" json:"base_url"`
    Models       json.RawMessage `gorm:"type:json;not null" json:"models"`   // JSON 字段
    ChannelGroup string          `gorm:"type:varchar(64);not null" json:"channel_group"`
    Tag          string          `gorm:"type:varchar(200);not null" json:"tag"`
    Setting      json.RawMessage `gorm:"type:json" json:"setting,omitempty"` // 可为空 JSON
}

func (Channel) TableName() string {
    return "channels"
}