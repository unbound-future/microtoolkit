package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/common/db"
	"github.com/AnimateAIPlatform/animate-ai/dao"
	"github.com/AnimateAIPlatform/animate-ai/models"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

// ToolComponentService 工具组件服务
type ToolComponentService struct {
	db          *gorm.DB
	componentDAO *dao.ToolComponentDAO
}

// NewToolComponentService 创建工具组件服务
func NewToolComponentService() *ToolComponentService {
	return &ToolComponentService{
		db:           db.DB,
		componentDAO: dao.NewToolComponentDAOWithDB(db.DB),
	}
}

// NewToolComponentServiceWithDB 使用指定的数据库连接创建工具组件服务
func NewToolComponentServiceWithDB(db *gorm.DB) *ToolComponentService {
	return &ToolComponentService{
		db:           db,
		componentDAO: dao.NewToolComponentDAOWithDB(db),
	}
}

// CreateComponent 创建工具组件
func (s *ToolComponentService) CreateComponent(ctx context.Context, userID, name, description, componentType, assetID, serviceURL, paramDesc, cronExpression string) (*models.ToolComponent, error) {
	// 生成组件ID
	componentID := s.generateComponentID(userID, name, time.Now().Unix())

	component := &models.ToolComponent{
		UserID:      userID,
		ComponentID: componentID,
		Name:        name,
		Description: description,
		Type:        componentType,
	}

	// 根据类型设置相应字段
	if componentType == models.ToolComponentTypeAsset {
		if assetID == "" {
			return nil, fmt.Errorf("asset ID is required for asset component")
		}
		component.AssetID = &assetID
	} else if componentType == models.ToolComponentTypeService {
		if serviceURL == "" {
			return nil, fmt.Errorf("service URL is required for service component")
		}
		component.ServiceURL = &serviceURL
		if paramDesc != "" {
			component.ParamDesc = &paramDesc
		}
	} else if componentType == models.ToolComponentTypeTrigger {
		if cronExpression == "" {
			return nil, fmt.Errorf("cron expression is required for trigger component")
		}
		component.CronExpression = &cronExpression
	} else {
		return nil, fmt.Errorf("invalid component type: %s", componentType)
	}

	err := s.componentDAO.Create(component)
	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	hlog.CtxInfof(ctx, "Component created: componentID=%s, userID=%s, type=%s", componentID, userID, componentType)
	return component, nil
}

// UpdateComponent 更新工具组件
func (s *ToolComponentService) UpdateComponent(ctx context.Context, componentID, userID, name, description, assetID, serviceURL, paramDesc, cronExpression string) (*models.ToolComponent, error) {
	component, err := s.componentDAO.GetByComponentID(componentID)
	if err != nil {
		return nil, fmt.Errorf("component not found: %w", err)
	}

	// 验证组件属于当前用户
	if component.UserID != userID {
		return nil, fmt.Errorf("component does not belong to user")
	}

	// 更新基本信息
	component.Name = name
	component.Description = description

	// 根据类型更新相应字段
	if component.Type == models.ToolComponentTypeAsset {
		if assetID == "" {
			return nil, fmt.Errorf("asset ID is required for asset component")
		}
		component.AssetID = &assetID
	} else if component.Type == models.ToolComponentTypeService {
		if serviceURL == "" {
			return nil, fmt.Errorf("service URL is required for service component")
		}
		component.ServiceURL = &serviceURL
		if paramDesc != "" {
			component.ParamDesc = &paramDesc
		}
	} else if component.Type == models.ToolComponentTypeTrigger {
		if cronExpression == "" {
			return nil, fmt.Errorf("cron expression is required for trigger component")
		}
		component.CronExpression = &cronExpression
	}

	err = s.componentDAO.Update(component)
	if err != nil {
		return nil, fmt.Errorf("failed to update component: %w", err)
	}

	hlog.CtxInfof(ctx, "Component updated: componentID=%s, userID=%s", componentID, userID)
	return component, nil
}

// DeleteComponent 删除工具组件
func (s *ToolComponentService) DeleteComponent(ctx context.Context, componentID, userID string) error {
	component, err := s.componentDAO.GetByComponentID(componentID)
	if err != nil {
		return fmt.Errorf("component not found: %w", err)
	}

	// 验证组件属于当前用户
	if component.UserID != userID {
		return fmt.Errorf("component does not belong to user")
	}

	err = s.componentDAO.Delete(component)
	if err != nil {
		return fmt.Errorf("failed to delete component: %w", err)
	}

	hlog.CtxInfof(ctx, "Component deleted: componentID=%s, userID=%s", componentID, userID)
	return nil
}

// ListComponents 列出用户的所有工具组件
func (s *ToolComponentService) ListComponents(ctx context.Context, userID string) ([]models.ToolComponent, error) {
	components, err := s.componentDAO.ListByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}
	return components, nil
}

// GetComponent 根据组件ID获取工具组件
func (s *ToolComponentService) GetComponent(ctx context.Context, componentID string) (*models.ToolComponent, error) {
	component, err := s.componentDAO.GetByComponentID(componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get component: %w", err)
	}
	return component, nil
}

// 辅助方法

// generateComponentID 生成组件ID
func (s *ToolComponentService) generateComponentID(userID, name string, timestamp int64) string {
	data := fmt.Sprintf("%s_%s_%d", userID, name, timestamp)
	hash := md5.Sum([]byte(data))
	hashStr := hex.EncodeToString(hash[:])
	// 截取前50个字符作为组件ID
	if len(hashStr) > 50 {
		hashStr = hashStr[:50]
	}
	return hashStr
}
