package dao

import (
    "context"

    "github.com/AnimateAIPlatform/animate-ai/common/db"
    "github.com/AnimateAIPlatform/animate-ai/models"
)

// CreatePurchaseOrder inserts a new purchase order into the database.
func CreatePurchaseOrder(ctx context.Context, order *models.PurchaseOrder) error {
    return db.DB.WithContext(ctx).Create(order).Error
}

// ListPurchaseOrdersByUser returns all purchase orders for a specific user.
func ListPurchaseOrdersByUser(ctx context.Context, userID string) ([]models.PurchaseOrder, error) {
    var orders []models.PurchaseOrder
    err := db.DB.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error
    return orders, err
}

// ListAllPurchaseOrders returns all purchase orders (admin use).
func ListAllPurchaseOrders(ctx context.Context) ([]models.PurchaseOrder, error) {
    var orders []models.PurchaseOrder
    err := db.DB.WithContext(ctx).Order("created_at desc").Find(&orders).Error
    return orders, err
}

// ApprovePurchaseOrder sets the status to approved and writes account info.
func ApprovePurchaseOrder(ctx context.Context, id uint, adminID string, accountInfo string) error {
    return db.DB.WithContext(ctx).Model(&models.PurchaseOrder{}).Where("id = ?", id).Updates(map[string]interface{}{
        "status":       "approved",
        "admin_id":     adminID,
        "account_info": accountInfo,
    }).Error
}
