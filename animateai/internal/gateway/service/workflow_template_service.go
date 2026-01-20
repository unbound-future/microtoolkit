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

// WorkflowTemplateService 工作流模版服务
type WorkflowTemplateService struct {
	db          *gorm.DB
	templateDAO *dao.WorkflowTemplateDAO
}

// NewWorkflowTemplateService 创建工作流模版服务
func NewWorkflowTemplateService() *WorkflowTemplateService {
	return &WorkflowTemplateService{
		db:          db.DB,
		templateDAO: dao.NewWorkflowTemplateDAOWithDB(db.DB),
	}
}

// NewWorkflowTemplateServiceWithDB 使用指定的数据库连接创建工作流模版服务
func NewWorkflowTemplateServiceWithDB(db *gorm.DB) *WorkflowTemplateService {
	return &WorkflowTemplateService{
		db:          db,
		templateDAO: dao.NewWorkflowTemplateDAOWithDB(db),
	}
}

// generateTemplateID 生成唯一的模版ID
func (s *WorkflowTemplateService) generateTemplateID(userID, name string, timestamp int64) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s_%s_%d", userID, name, timestamp)))
	return hex.EncodeToString(hash[:])
}

// generateUniqueID 生成唯一的ID（用于资产ID）
func (s *WorkflowTemplateService) generateUniqueID() string {
	timestamp := time.Now().UnixNano()
	// 使用时间戳和随机数生成唯一ID
	hash := md5.Sum([]byte(fmt.Sprintf("%d_%d_%d", timestamp, time.Now().Unix(), time.Now().UnixNano())))
	return hex.EncodeToString(hash[:16]) // 取前16个字符作为ID
}

// CreateWorkflowTemplate 创建工作流模版
func (s *WorkflowTemplateService) CreateWorkflowTemplate(ctx context.Context, userID, name, description, assetID string, templateData interface{}) (*models.WorkflowTemplate, error) {
	// 验证输入
	if name == "" {
		return nil, fmt.Errorf("template name is required")
	}

	// 将 templateData 转换为 JSON 字符串
	templateDataJSON, err := json.Marshal(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template data: %w", err)
	}

	// 生成唯一的模版ID
	templateID := s.generateTemplateID(userID, name, time.Now().UnixNano())

	// 如果没有提供资产ID，生成一个唯一的资产ID
	if assetID == "" {
		assetID = s.generateUniqueID()
		hlog.CtxInfof(ctx, "Generated asset ID for new template: %s", assetID)
	}

	template := &models.WorkflowTemplate{
		UserID:       userID,
		TemplateID:   templateID,
		Name:         name,
		Description:  description,
		AssetID:      assetID,
		TemplateData: string(templateDataJSON),
	}

	err = s.templateDAO.Create(template)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow template: %w", err)
	}

	hlog.CtxInfof(ctx, "Workflow template created: templateID=%s, userID=%s, name=%s", templateID, userID, name)
	return template, nil
}

// GetWorkflowTemplate 获取工作流模版详情
func (s *WorkflowTemplateService) GetWorkflowTemplate(ctx context.Context, templateID, userID string) (*models.WorkflowTemplate, error) {
	template, err := s.templateDAO.GetByTemplateID(templateID)
	if err != nil {
		return nil, fmt.Errorf("workflow template not found: %w", err)
	}
	if template.UserID != userID {
		return nil, fmt.Errorf("workflow template does not belong to user")
	}
	return template, nil
}

// UpdateWorkflowTemplate 更新工作流模版信息
func (s *WorkflowTemplateService) UpdateWorkflowTemplate(ctx context.Context, templateID, userID, name, description, assetID string, templateData interface{}) (*models.WorkflowTemplate, error) {
	template, err := s.templateDAO.GetByTemplateID(templateID)
	if err != nil {
		return nil, fmt.Errorf("workflow template not found: %w", err)
	}
	if template.UserID != userID {
		return nil, fmt.Errorf("workflow template does not belong to user")
	}

	// 验证输入
	if name == "" {
		return nil, fmt.Errorf("template name is required")
	}

	// 将 templateData 转换为 JSON 字符串
	templateDataJSON, err := json.Marshal(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template data: %w", err)
	}

	template.Name = name
	template.Description = description
	
	// 更新资产ID：如果提供了新的资产ID则使用，否则保持原有值，如果原来没有则生成新的
	if assetID != "" {
		template.AssetID = assetID
	} else if template.AssetID == "" {
		template.AssetID = s.generateUniqueID()
		hlog.CtxInfof(ctx, "Generated asset ID for existing template: %s", template.AssetID)
	}
	
	template.TemplateData = string(templateDataJSON)

	err = s.templateDAO.Update(template)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow template: %w", err)
	}

	hlog.CtxInfof(ctx, "Workflow template updated: templateID=%s, userID=%s", templateID, userID)
	return template, nil
}

// DeleteWorkflowTemplate 删除工作流模版
func (s *WorkflowTemplateService) DeleteWorkflowTemplate(ctx context.Context, templateID, userID string) error {
	template, err := s.templateDAO.GetByTemplateID(templateID)
	if err != nil {
		return fmt.Errorf("workflow template not found: %w", err)
	}
	if template.UserID != userID {
		return fmt.Errorf("workflow template does not belong to user")
	}

	err = s.templateDAO.Delete(template)
	if err != nil {
		return fmt.Errorf("failed to delete workflow template: %w", err)
	}

	hlog.CtxInfof(ctx, "Workflow template deleted: templateID=%s, userID=%s", templateID, userID)
	return nil
}

// ListWorkflowTemplates 列出用户的所有工作流模版
func (s *WorkflowTemplateService) ListWorkflowTemplates(ctx context.Context, userID string) ([]models.WorkflowTemplate, error) {
	templates, err := s.templateDAO.ListByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow templates: %w", err)
	}
	return templates, nil
}
