package dao

import (
	"github.com/AnimateAIPlatform/animate-ai/models"

	"gorm.io/gorm"
)

// WorkflowTemplateDAO 工作流模版 DAO
type WorkflowTemplateDAO struct {
	db *gorm.DB
}

// NewWorkflowTemplateDAOWithDB 使用指定的数据库连接创建工作流模版 DAO
func NewWorkflowTemplateDAOWithDB(db *gorm.DB) *WorkflowTemplateDAO {
	return &WorkflowTemplateDAO{db: db}
}

// Create 插入新工作流模版
func (dao *WorkflowTemplateDAO) Create(template *models.WorkflowTemplate) error {
	return dao.db.Create(template).Error
}

// Update 更新工作流模版
func (dao *WorkflowTemplateDAO) Update(template *models.WorkflowTemplate) error {
	return dao.db.Save(template).Error
}

// Delete 软删除工作流模版
func (dao *WorkflowTemplateDAO) Delete(template *models.WorkflowTemplate) error {
	return dao.db.Delete(template).Error
}

// GetByID 根据ID查询工作流模版
func (dao *WorkflowTemplateDAO) GetByID(id uint) (*models.WorkflowTemplate, error) {
	var template models.WorkflowTemplate
	err := dao.db.Where("id = ? AND deleted_at IS NULL", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetByTemplateID 根据模版ID查询工作流模版
func (dao *WorkflowTemplateDAO) GetByTemplateID(templateID string) (*models.WorkflowTemplate, error) {
	var template models.WorkflowTemplate
	err := dao.db.Where("template_id = ? AND deleted_at IS NULL", templateID).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// ListByUserID 查询指定用户的所有工作流模版
func (dao *WorkflowTemplateDAO) ListByUserID(userID string) ([]models.WorkflowTemplate, error) {
	var templates []models.WorkflowTemplate
	err := dao.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&templates).Error
	return templates, err
}

// SearchByUserIDAndName 根据用户ID和名称搜索工作流模版
func (dao *WorkflowTemplateDAO) SearchByUserIDAndName(userID, name string) ([]models.WorkflowTemplate, error) {
	var templates []models.WorkflowTemplate
	err := dao.db.Where("user_id = ? AND name LIKE ? AND deleted_at IS NULL", userID, "%"+name+"%").Find(&templates).Error
	return templates, err
}
