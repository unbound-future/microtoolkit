package models

import (
	"gorm.io/gorm"
)

// User 用户表
type User struct {
	gorm.Model
	UserName        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"user_name"`        // 用户名（唯一）
	Password        string `gorm:"type:varchar(255);not null" json:"-"`                            // 密码（不返回给前端）
	Email           string `gorm:"type:varchar(255);index" json:"email,omitempty"`                 // 邮箱（可选）
	Status          int    `gorm:"default:1;index" json:"status"`                                    // 状态：1=正常, 0=禁用
	Name            string `gorm:"type:varchar(100)" json:"name,omitempty"`                         // 昵称/显示名称（可选，为空时使用UserName）
	Avatar          string `gorm:"type:varchar(500)" json:"avatar,omitempty"`                       // 头像URL
	Job             string `gorm:"type:varchar(50)" json:"job,omitempty"`                          // 职位代码
	JobName         string `gorm:"type:varchar(100)" json:"job_name,omitempty"`                    // 职位名称
	Organization    string `gorm:"type:varchar(50)" json:"organization,omitempty"`                 // 组织代码
	OrganizationName string `gorm:"type:varchar(100)" json:"organization_name,omitempty"`          // 组织名称
	Location        string `gorm:"type:varchar(50)" json:"location,omitempty"`                    // 位置代码
	LocationName    string `gorm:"type:varchar(100)" json:"location_name,omitempty"`               // 位置名称
	Introduction    string `gorm:"type:text" json:"introduction,omitempty"`                        // 个人简介
	PersonalWebsite string `gorm:"type:varchar(500)" json:"personal_website,omitempty"`           // 个人网站
	Verified        bool   `gorm:"default:false" json:"verified,omitempty"`                         // 是否验证
	PhoneNumber     string `gorm:"type:varchar(20)" json:"phone_number,omitempty"`                 // 电话号码
	AccountID       string `gorm:"type:varchar(50);uniqueIndex" json:"account_id,omitempty"`             // 账户ID（唯一）
	Address         string `gorm:"type:varchar(500)" json:"address,omitempty"`                    // 具体地址
	RangeArea       string `gorm:"type:varchar(100)" json:"range_area,omitempty"`                    // 国家/地区
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

