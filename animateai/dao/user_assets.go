package dao

import (
	"github.com/AnimateAIPlatform/animate-ai/models"

	"gorm.io/gorm"
)

// UserAssetDAO 用户资产 DAO
type UserAssetDAO struct {
	db *gorm.DB
}

// NewUserAssetDAOWithDB 使用指定的数据库连接创建用户资产 DAO
func NewUserAssetDAOWithDB(db *gorm.DB) *UserAssetDAO {
	return &UserAssetDAO{db: db}
}

// Create 插入新用户资产
func (dao *UserAssetDAO) Create(asset *models.UserAsset) error {
	return dao.db.Create(asset).Error
}

// Update 更新用户资产
func (dao *UserAssetDAO) Update(asset *models.UserAsset) error {
	return dao.db.Save(asset).Error
}

// Delete 软删除用户资产
func (dao *UserAssetDAO) Delete(asset *models.UserAsset) error {
	return dao.db.Delete(asset).Error
}

// GetByID 根据ID查询用户资产
func (dao *UserAssetDAO) GetByID(id uint) (*models.UserAsset, error) {
	var asset models.UserAsset
	err := dao.db.Where("id = ? AND deleted_at IS NULL", id).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// GetByAssetID 根据资产ID查询用户资产
func (dao *UserAssetDAO) GetByAssetID(assetID string) (*models.UserAsset, error) {
	var asset models.UserAsset
	err := dao.db.Where("asset_id = ? AND deleted_at IS NULL", assetID).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// ListByUserID 查询指定用户的所有资产
func (dao *UserAssetDAO) ListByUserID(userID string) ([]models.UserAsset, error) {
	var assets []models.UserAsset
	err := dao.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&assets).Error
	return assets, err
}

// ListByUserIDAndType 根据用户ID和类型查询资产
func (dao *UserAssetDAO) ListByUserIDAndType(userID, assetType string) ([]models.UserAsset, error) {
	var assets []models.UserAsset
	err := dao.db.Where("user_id = ? AND type = ? AND deleted_at IS NULL", userID, assetType).Find(&assets).Error
	return assets, err
}

// ListAll 查询所有用户资产（未删除的）
func (dao *UserAssetDAO) ListAll() ([]models.UserAsset, error) {
	var assets []models.UserAsset
	err := dao.db.Where("deleted_at IS NULL").Find(&assets).Error
	return assets, err
}

// SearchByUserIDAndName 根据用户ID和名称搜索资产
func (dao *UserAssetDAO) SearchByUserIDAndName(userID, name string) ([]models.UserAsset, error) {
	var assets []models.UserAsset
	err := dao.db.Where("user_id = ? AND name LIKE ? AND deleted_at IS NULL", userID, "%"+name+"%").Find(&assets).Error
	return assets, err
}
