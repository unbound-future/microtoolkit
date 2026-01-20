package dao

import (
	"github.com/AnimateAIPlatform/animate-ai/models"

	"gorm.io/gorm"
)

// ToolComponentDAO 工具组件 DAO
type ToolComponentDAO struct {
	db *gorm.DB
}

// NewToolComponentDAOWithDB 使用指定的数据库连接创建工具组件 DAO
func NewToolComponentDAOWithDB(db *gorm.DB) *ToolComponentDAO {
	return &ToolComponentDAO{db: db}
}

// Create 插入新工具组件
func (dao *ToolComponentDAO) Create(component *models.ToolComponent) error {
	return dao.db.Create(component).Error
}

// Update 更新工具组件
func (dao *ToolComponentDAO) Update(component *models.ToolComponent) error {
	return dao.db.Save(component).Error
}

// Delete 软删除工具组件
func (dao *ToolComponentDAO) Delete(component *models.ToolComponent) error {
	return dao.db.Delete(component).Error
}

// GetByID 根据ID查询工具组件
func (dao *ToolComponentDAO) GetByID(id uint) (*models.ToolComponent, error) {
	var component models.ToolComponent
	err := dao.db.Where("id = ? AND deleted_at IS NULL", id).First(&component).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// GetByComponentID 根据组件ID查询工具组件
func (dao *ToolComponentDAO) GetByComponentID(componentID string) (*models.ToolComponent, error) {
	var component models.ToolComponent
	err := dao.db.Where("component_id = ? AND deleted_at IS NULL", componentID).First(&component).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// ListByUserID 查询指定用户的所有工具组件
func (dao *ToolComponentDAO) ListByUserID(userID string) ([]models.ToolComponent, error) {
	var components []models.ToolComponent
	err := dao.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&components).Error
	return components, err
}

// ListByUserIDAndType 根据用户ID和类型查询工具组件
func (dao *ToolComponentDAO) ListByUserIDAndType(userID, componentType string) ([]models.ToolComponent, error) {
	var components []models.ToolComponent
	err := dao.db.Where("user_id = ? AND type = ? AND deleted_at IS NULL", userID, componentType).Find(&components).Error
	return components, err
}

// SearchByUserIDAndName 根据用户ID和名称搜索工具组件
func (dao *ToolComponentDAO) SearchByUserIDAndName(userID, name string) ([]models.ToolComponent, error) {
	var components []models.ToolComponent
	err := dao.db.Where("user_id = ? AND name LIKE ? AND deleted_at IS NULL", userID, "%"+name+"%").Find(&components).Error
	return components, err
}
