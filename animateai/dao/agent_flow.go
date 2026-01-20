package dao

import (
	"github.com/AnimateAIPlatform/animate-ai/models"

	"gorm.io/gorm"
)

// AgentFlowDAO 工作流 DAO
type AgentFlowDAO struct {
	db *gorm.DB
}

// NewAgentFlowDAOWithDB 使用指定的数据库连接创建工作流 DAO
func NewAgentFlowDAOWithDB(db *gorm.DB) *AgentFlowDAO {
	return &AgentFlowDAO{db: db}
}

// Create 插入新工作流
func (dao *AgentFlowDAO) Create(flow *models.AgentFlow) error {
	return dao.db.Create(flow).Error
}

// Update 更新工作流
func (dao *AgentFlowDAO) Update(flow *models.AgentFlow) error {
	return dao.db.Save(flow).Error
}

// Delete 软删除工作流
func (dao *AgentFlowDAO) Delete(flow *models.AgentFlow) error {
	return dao.db.Delete(flow).Error
}

// GetByID 根据ID查询工作流
func (dao *AgentFlowDAO) GetByID(id uint) (*models.AgentFlow, error) {
	var flow models.AgentFlow
	err := dao.db.Where("id = ? AND deleted_at IS NULL", id).First(&flow).Error
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

// GetByFlowID 根据工作流ID查询工作流
func (dao *AgentFlowDAO) GetByFlowID(flowID string) (*models.AgentFlow, error) {
	var flow models.AgentFlow
	err := dao.db.Where("flow_id = ? AND deleted_at IS NULL", flowID).First(&flow).Error
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

// ListByUserID 查询指定用户的所有工作流
func (dao *AgentFlowDAO) ListByUserID(userID string) ([]models.AgentFlow, error) {
	var flows []models.AgentFlow
	err := dao.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&flows).Error
	return flows, err
}

// SearchByUserIDAndName 根据用户ID和名称搜索工作流
func (dao *AgentFlowDAO) SearchByUserIDAndName(userID, name string) ([]models.AgentFlow, error) {
	var flows []models.AgentFlow
	err := dao.db.Where("user_id = ? AND name LIKE ? AND deleted_at IS NULL", userID, "%"+name+"%").Find(&flows).Error
	return flows, err
}



