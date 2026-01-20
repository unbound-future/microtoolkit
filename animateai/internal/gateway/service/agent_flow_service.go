package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/common/db"
	"github.com/AnimateAIPlatform/animate-ai/dao"
	"github.com/AnimateAIPlatform/animate-ai/models"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

// AgentFlowService 工作流服务
type AgentFlowService struct {
	db          *gorm.DB
	agentFlowDAO *dao.AgentFlowDAO
}

// NewAgentFlowService 创建工作流服务
func NewAgentFlowService() *AgentFlowService {
	return &AgentFlowService{
		db:           db.DB,
		agentFlowDAO: dao.NewAgentFlowDAOWithDB(db.DB),
	}
}

// NewAgentFlowServiceWithDB 使用指定的数据库连接创建工作流服务
func NewAgentFlowServiceWithDB(db *gorm.DB) *AgentFlowService {
	return &AgentFlowService{
		db:           db,
		agentFlowDAO: dao.NewAgentFlowDAOWithDB(db),
	}
}

// generateFlowID 生成唯一的工作流ID
func (s *AgentFlowService) generateFlowID(userID, name string, timestamp int64) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s_%s_%d", userID, name, timestamp)))
	return hex.EncodeToString(hash[:])
}

// generateUniqueID 生成唯一的ID（用于模版ID和资产ID）
func (s *AgentFlowService) generateUniqueID() string {
	timestamp := time.Now().UnixNano()
	// 使用时间戳和随机数生成唯一ID
	hash := md5.Sum([]byte(fmt.Sprintf("%d_%d_%d", timestamp, time.Now().Unix(), time.Now().UnixNano())))
	return hex.EncodeToString(hash[:16]) // 取前16个字符作为ID
}

// CreateAgentFlow 创建工作流
func (s *AgentFlowService) CreateAgentFlow(ctx context.Context, userID, name, assetID, templateID string, flowData interface{}) (*models.AgentFlow, error) {
	// 验证输入
	if name == "" {
		return nil, fmt.Errorf("flow name is required")
	}

	// 将 flowData 转换为 JSON 字符串
	flowDataJSON, err := json.Marshal(flowData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow data: %w", err)
	}

	// 生成唯一的工作流ID
	flowID := s.generateFlowID(userID, name, time.Now().UnixNano())

	// 如果没有提供资产ID，生成一个唯一的资产ID
	if assetID == "" {
		assetID = s.generateUniqueID()
		hlog.CtxInfof(ctx, "Generated asset ID for new workflow: %s", assetID)
	}

	// 如果没有提供模版ID，生成一个唯一的模版ID
	if templateID == "" {
		templateID = s.generateUniqueID()
		hlog.CtxInfof(ctx, "Generated template ID for new workflow: %s", templateID)
	}

	flow := &models.AgentFlow{
		UserID:     userID,
		FlowID:     flowID,
		Name:       name,
		AssetID:    assetID,
		TemplateID: templateID,
		FlowData:   string(flowDataJSON),
	}

	err = s.agentFlowDAO.Create(flow)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent flow: %w", err)
	}

	hlog.CtxInfof(ctx, "Agent flow created: flowID=%s, userID=%s, name=%s", flowID, userID, name)
	return flow, nil
}

// GetAgentFlow 获取工作流详情
func (s *AgentFlowService) GetAgentFlow(ctx context.Context, flowID, userID string) (*models.AgentFlow, error) {
	flow, err := s.agentFlowDAO.GetByFlowID(flowID)
	if err != nil {
		return nil, fmt.Errorf("agent flow not found: %w", err)
	}
	if flow.UserID != userID {
		return nil, fmt.Errorf("agent flow does not belong to user")
	}
	return flow, nil
}

// UpdateAgentFlow 更新工作流信息
func (s *AgentFlowService) UpdateAgentFlow(ctx context.Context, flowID, userID, name, assetID, templateID string, flowData interface{}) (*models.AgentFlow, error) {
	flow, err := s.agentFlowDAO.GetByFlowID(flowID)
	if err != nil {
		return nil, fmt.Errorf("agent flow not found: %w", err)
	}
	if flow.UserID != userID {
		return nil, fmt.Errorf("agent flow does not belong to user")
	}

	// 验证输入
	if name == "" {
		return nil, fmt.Errorf("flow name is required")
	}

	// 将 flowData 转换为 JSON 字符串
	flowDataJSON, err := json.Marshal(flowData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow data: %w", err)
	}

	flow.Name = name
	flow.AssetID = assetID
	flow.TemplateID = templateID
	flow.FlowData = string(flowDataJSON)

	err = s.agentFlowDAO.Update(flow)
	if err != nil {
		return nil, fmt.Errorf("failed to update agent flow: %w", err)
	}

	hlog.CtxInfof(ctx, "Agent flow updated: flowID=%s, userID=%s", flowID, userID)
	return flow, nil
}

// DeleteAgentFlow 删除工作流
func (s *AgentFlowService) DeleteAgentFlow(ctx context.Context, flowID, userID string) error {
	flow, err := s.agentFlowDAO.GetByFlowID(flowID)
	if err != nil {
		return fmt.Errorf("agent flow not found: %w", err)
	}
	if flow.UserID != userID {
		return fmt.Errorf("agent flow does not belong to user")
	}

	err = s.agentFlowDAO.Delete(flow)
	if err != nil {
		return fmt.Errorf("failed to delete agent flow: %w", err)
	}

	hlog.CtxInfof(ctx, "Agent flow deleted: flowID=%s, userID=%s", flowID, userID)
	return nil
}

// ListAgentFlows 列出用户的所有工作流
func (s *AgentFlowService) ListAgentFlows(ctx context.Context, userID string) ([]models.AgentFlow, error) {
	flows, err := s.agentFlowDAO.ListByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent flows: %w", err)
	}
	return flows, nil
}



