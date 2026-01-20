package model

import "time"

// ChannelModelSchedule 对应 channel_model_schedule 表
type ChannelModelSchedule struct {
    ID        int64   `gorm:"primaryKey;autoIncrement" json:"id"`                    // 自增主键
    Type      string  `gorm:"type:varchar(50);not null" json:"type"`               // 模型类型
    ChannelID int64   `gorm:"not null" json:"channel_id"`                           // 关联 channels.id
    ModelName string  `gorm:"type:varchar(100);not null" json:"model_name"`        // 关联 model_pricing.model_name
    Priority  uint    `gorm:"default:0" json:"priority"`                             // 调度优先级
    Status    uint8   `gorm:"default:1" json:"status"`                              // 调度状态：1=有效,0=无效
    Weight    uint    `gorm:"default:0" json:"weight"`                              // 权重
    Remark    *string `gorm:"type:varchar(255)" json:"remark,omitempty"`           // 备注，可为空
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                   // 可选：记录创建时间
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                   // 可选：记录更新时间
}

func (ChannelModelSchedule) TableName() string {
    return "channel_model_schedule"
}