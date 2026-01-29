package models

import "gorm.io/gorm"

// PurchaseOrder represents a user purchase request for AI-VIP account
// Plan values: "vip_7d", "vip_30d", "vip_90d"
// Status values: "pending", "approved", "rejected"
// If ApprovedAt is non-zero, order is approved.
// AdminID records who approved.
// AccountInfo may contain delivered account credentials or message shown to user.

type PurchaseOrder struct {
    gorm.Model
    UserID     string `gorm:"type:varchar(100);not null;index" json:"user_id"`
    Plan       string `gorm:"type:varchar(20);not null" json:"plan"`
    Status     string `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
    AccountInfo string `gorm:"type:text" json:"account_info"`
    AdminID    string `gorm:"type:varchar(100)" json:"admin_id"`
}

func (PurchaseOrder) TableName() string {
    return "purchase_orders"
}