package service

import (
	"context"
	"fmt"

	"github.com/AnimateAIPlatform/animate-ai/common/cloud/oss"
	"github.com/AnimateAIPlatform/animate-ai/common/db"
	"github.com/AnimateAIPlatform/animate-ai/models"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

// OSSService 对象存储服务
type OSSService struct {
	db *gorm.DB
}

// NewOSSService 创建对象存储服务
func NewOSSService() *OSSService {
	return &OSSService{
		db: db.DB,
	}
}

// NewOSSServiceWithDB 使用指定的数据库连接创建对象存储服务
func NewOSSServiceWithDB(db *gorm.DB) *OSSService {
	return &OSSService{
		db: db,
	}
}

// GetStorageConfigByID 根据ID获取对象存储配置
func (s *OSSService) GetStorageConfigByID(ctx context.Context, id uint) (*models.ObjectStorageConfig, error) {
	var config models.ObjectStorageConfig
	err := s.db.Where("id = ? AND deleted_at IS NULL AND status = ?", id, 1).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("object storage config not found: id=%d", id)
		}
		return nil, fmt.Errorf("failed to get object storage config: %w", err)
	}
	return &config, nil
}

// GetOSSClientByConfigID 根据配置ID获取OSS客户端
func (s *OSSService) GetOSSClientByConfigID(ctx context.Context, configID uint) (oss.Client, *models.ObjectStorageConfig, error) {
	// 获取配置
	config, err := s.GetStorageConfigByID(ctx, configID)
	if err != nil {
		return nil, nil, err
	}

	// 转换为OSS配置
	ossConfig := oss.Config{
		Type:      config.Type,
		Endpoint:  config.Endpoint,
		Bucket:    config.Bucket,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Region:    config.Region,
		BaseURL:   config.BaseURL,
	}

	// 验证配置完整性
	if ossConfig.Type == "" {
		return nil, nil, fmt.Errorf("storage type is empty")
	}
	if ossConfig.Bucket == "" {
		return nil, nil, fmt.Errorf("bucket name is empty")
	}
	if ossConfig.AccessKey == "" {
		return nil, nil, fmt.Errorf("access key is empty")
	}
	if ossConfig.SecretKey == "" {
		return nil, nil, fmt.Errorf("secret key is empty")
	}

	// 创建OSS客户端
	client, err := oss.NewClient(ossConfig)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to create OSS client: type=%s, bucket=%s, error=%v", ossConfig.Type, ossConfig.Bucket, err)
		return nil, nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	// 验证客户端创建成功
	if client == nil {
		return nil, nil, fmt.Errorf("OSS client is nil after creation")
	}

	hlog.CtxInfof(ctx, "OSS client created successfully: type=%s, bucket=%s, region=%s", ossConfig.Type, ossConfig.Bucket, ossConfig.Region)

	return client, config, nil
}

// ListObjects 列出对象存储中的对象
func (s *OSSService) ListObjects(ctx context.Context, configID uint, prefix, marker string, maxKeys int) ([]oss.ObjectInfo, error) {
	// 获取OSS客户端
	client, config, err := s.GetOSSClientByConfigID(ctx, configID)
	if err != nil {
		return nil, err
	}

	// 列出对象
	objects, err := client.ListObjects(ctx, config.Bucket, prefix, marker, maxKeys)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to list objects: configID=%d, bucket=%s, error=%v", configID, config.Bucket, err)
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	hlog.CtxInfof(ctx, "Listed %d objects from bucket %s (configID=%d)", len(objects), config.Bucket, configID)
	return objects, nil
}
