package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/common/db"
	"github.com/AnimateAIPlatform/animate-ai/dao"
	"github.com/AnimateAIPlatform/animate-ai/models"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

// AssetService 资产管理服务
type AssetService struct {
	db          *gorm.DB
	ossService  *OSSService
	assetDAO    *dao.UserAssetDAO
}

// NewAssetService 创建资产管理服务
func NewAssetService() *AssetService {
	return &AssetService{
		db:         db.DB,
		ossService: NewOSSService(),
		assetDAO:   dao.NewUserAssetDAOWithDB(db.DB),
	}
}

// NewAssetServiceWithDB 使用指定的数据库连接创建资产管理服务
func NewAssetServiceWithDB(db *gorm.DB) *AssetService {
	return &AssetService{
		db:         db,
		ossService: NewOSSServiceWithDB(db),
		assetDAO:   dao.NewUserAssetDAOWithDB(db),
	}
}

// GetDefaultStorageConfig 获取默认的存储配置
func (s *AssetService) GetDefaultStorageConfig(ctx context.Context) (*models.ObjectStorageConfig, error) {
	var config models.ObjectStorageConfig
	err := s.db.Where("is_default = ? AND status = ? AND deleted_at IS NULL", true, 1).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有默认配置，尝试获取第一个启用的COS配置
			hlog.CtxInfof(ctx, "No default storage config found, trying to get first COS config")
			err = s.db.Where("type = ? AND status = ? AND deleted_at IS NULL", "cos", 1).First(&config).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					hlog.CtxErrorf(ctx, "No COS storage config found in database")
					return nil, fmt.Errorf("no storage config found: please configure object storage (COS) in database table 'object_storage_configs'")
				}
				return nil, fmt.Errorf("failed to get COS storage config: %w", err)
			}
			hlog.CtxInfof(ctx, "Using COS config: id=%d, name=%s, bucket=%s", config.ID, config.Name, config.Bucket)
		} else {
			hlog.CtxErrorf(ctx, "Failed to query default storage config: %v", err)
			return nil, fmt.Errorf("failed to get default storage config: %w", err)
		}
	} else {
		hlog.CtxInfof(ctx, "Using default storage config: id=%d, name=%s, bucket=%s", config.ID, config.Name, config.Bucket)
	}
	return &config, nil
}

// UploadAsset 上传资产到对象存储并保存到数据库
func (s *AssetService) UploadAsset(ctx context.Context, userID string, name, description string, file io.Reader, fileName string, contentType string, fileSize int64) (*models.UserAsset, error) {
	// 1. 获取默认存储配置
	storageConfig, err := s.GetDefaultStorageConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage config: %w", err)
	}

	// 验证存储配置
	if storageConfig == nil {
		return nil, fmt.Errorf("storage config is nil")
	}

	// 2. 生成对象键名（objectKey）
	// 格式：assets/{userID}/{timestamp}_{hash}_{filename}
	timestamp := time.Now().Unix()
	hash := s.generateFileHash(fileName, userID, timestamp)
	ext := filepath.Ext(fileName)
	if ext == "" {
		// 如果没有扩展名，尝试从contentType推断
		ext = ".bin"
	}
	objectKey := fmt.Sprintf("assets/%s/%d_%s%s", userID, timestamp, hash, ext)

	hlog.CtxInfof(ctx, "Uploading asset: userID=%s, objectKey=%s, fileName=%s, size=%d, contentType=%s", userID, objectKey, fileName, fileSize, contentType)

	// 3. 获取OSS客户端并上传文件
	ossClient, config, err := s.ossService.GetOSSClientByConfigID(ctx, storageConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get OSS client: %w", err)
	}

	// 安全检查
	if ossClient == nil {
		return nil, fmt.Errorf("OSS client is nil")
	}
	if config == nil {
		return nil, fmt.Errorf("storage config is nil")
	}
	if config.Bucket == "" {
		return nil, fmt.Errorf("bucket name is empty in storage config")
	}

	hlog.CtxInfof(ctx, "OSS client created: bucket=%s, region=%s, endpoint=%s", config.Bucket, config.Region, config.Endpoint)

	// 上传文件到COS
	hlog.CtxInfof(ctx, "Starting upload to COS: bucket=%s, objectKey=%s, contentType=%s, fileSize=%d", config.Bucket, objectKey, contentType, fileSize)
	storageURL, err := ossClient.Upload(ctx, config.Bucket, objectKey, file, contentType)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to upload to COS: bucket=%s, objectKey=%s, error=%v", config.Bucket, objectKey, err)
		return nil, fmt.Errorf("failed to upload file to COS: %w", err)
	}
	hlog.CtxInfof(ctx, "Successfully uploaded to COS: storageURL=%s", storageURL)

	// 验证返回的URL不为空
	if storageURL == "" {
		return nil, fmt.Errorf("storage URL is empty after upload")
	}

	hlog.CtxInfof(ctx, "File uploaded to COS: storageURL=%s", storageURL)

	// 4. 检测资产类型
	assetType := s.detectAssetType(fileName, contentType)

	// 5. 生成资产ID
	assetID := s.generateAssetID(userID, timestamp)

	// 6. 保存到数据库
	asset := &models.UserAsset{
		UserID:          userID,
		AssetID:         assetID,
		Name:            name,
		Description:     description,
		URL:             storageURL,
		Source:          "file",
		Type:            assetType,
		Size:            &fileSize,
		MimeType:        contentType,
		StorageConfigID: &storageConfig.ID,
		StorageURL:      storageURL,
	}

	err = s.assetDAO.Create(asset)
	if err != nil {
		// 如果数据库保存失败，尝试删除已上传的文件
		hlog.CtxErrorf(ctx, "Failed to save asset to database, attempting to delete uploaded file: %v", err)
		if ossClient != nil && config != nil && config.Bucket != "" {
			_ = ossClient.Delete(ctx, config.Bucket, objectKey)
		}
		return nil, fmt.Errorf("failed to save asset to database: %w", err)
	}

	hlog.CtxInfof(ctx, "Asset saved successfully: assetID=%s, userID=%s", assetID, userID)

	return asset, nil
}

// AddAssetByURL 通过URL添加资产（不涉及文件上传）
func (s *AssetService) AddAssetByURL(ctx context.Context, userID, name, description, url string) (*models.UserAsset, error) {
	// 检测资产类型
	assetType := s.detectAssetTypeFromURL(url)

	// 生成资产ID
	assetID := s.generateAssetID(userID, time.Now().Unix())

	// 保存到数据库
	asset := &models.UserAsset{
		UserID:      userID,
		AssetID:     assetID,
		Name:        name,
		Description: description,
		URL:         url,
		Source:      "url",
		Type:        assetType,
	}

	err := s.assetDAO.Create(asset)
	if err != nil {
		return nil, fmt.Errorf("failed to save asset to database: %w", err)
	}

	hlog.CtxInfof(ctx, "Asset added by URL: assetID=%s, userID=%s, url=%s", assetID, userID, url)

	return asset, nil
}

// ListAssets 列出用户的所有资产
func (s *AssetService) ListAssets(ctx context.Context, userID string) ([]models.UserAsset, error) {
	assets, err := s.assetDAO.ListByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}
	return assets, nil
}

// GetAsset 根据资产ID获取资产
func (s *AssetService) GetAsset(ctx context.Context, assetID string) (*models.UserAsset, error) {
	asset, err := s.assetDAO.GetByAssetID(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}
	return asset, nil
}

// UpdateAsset 更新资产信息（名称、描述等，不涉及文件替换）
func (s *AssetService) UpdateAsset(ctx context.Context, assetID, userID, name, description, url string) (*models.UserAsset, error) {
	asset, err := s.assetDAO.GetByAssetID(assetID)
	if err != nil {
		return nil, fmt.Errorf("asset not found: %w", err)
	}

	// 验证资产属于当前用户
	if asset.UserID != userID {
		return nil, fmt.Errorf("asset does not belong to user")
	}

	// 更新资产信息
	asset.Name = name
	asset.Description = description
	if url != "" && asset.Source == "url" {
		asset.URL = url
		asset.Type = s.detectAssetTypeFromURL(url)
	}

	err = s.assetDAO.Update(asset)
	if err != nil {
		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	hlog.CtxInfof(ctx, "Asset updated: assetID=%s, userID=%s", assetID, userID)

	return asset, nil
}

// DeleteAsset 删除资产（同时删除对象存储中的文件）
func (s *AssetService) DeleteAsset(ctx context.Context, assetID, userID string) error {
	asset, err := s.assetDAO.GetByAssetID(assetID)
	if err != nil {
		return fmt.Errorf("asset not found: %w", err)
	}

	// 验证资产属于当前用户
	if asset.UserID != userID {
		return fmt.Errorf("asset does not belong to user")
	}

	// 如果是文件类型，需要从对象存储中删除
	if asset.Source == "file" && asset.StorageConfigID != nil {
		ossClient, storageConfig, err := s.ossService.GetOSSClientByConfigID(ctx, *asset.StorageConfigID)
		if err == nil {
			// 从 StorageURL 中提取 objectKey
			// StorageURL 格式可能是：https://bucket.cos.region.myqcloud.com/assets/userID/timestamp_hash.ext
			// 需要提取 objectKey 部分（assets/userID/timestamp_hash.ext）
			objectKey := s.extractObjectKeyFromURL(asset.StorageURL, storageConfig.BaseURL)
			if objectKey != "" {
				err = ossClient.Delete(ctx, storageConfig.Bucket, objectKey)
				if err != nil {
					hlog.CtxErrorf(ctx, "Failed to delete file from COS: %v", err)
					// 不阻止删除操作，继续删除数据库记录
				} else {
					hlog.CtxInfof(ctx, "File deleted from COS: objectKey=%s", objectKey)
				}
			}
		}
	}

	// 删除数据库记录（软删除）
	err = s.assetDAO.Delete(asset)
	if err != nil {
		return fmt.Errorf("failed to delete asset from database: %w", err)
	}

	hlog.CtxInfof(ctx, "Asset deleted: assetID=%s, userID=%s", assetID, userID)

	return nil
}

// GeneratePresignedURL 生成资产的预签名下载链接（1小时有效）
func (s *AssetService) GeneratePresignedURL(ctx context.Context, assetID, userID string) (string, error) {
	asset, err := s.assetDAO.GetByAssetID(assetID)
	if err != nil {
		return "", fmt.Errorf("asset not found: %w", err)
	}

	// 验证资产属于当前用户
	if asset.UserID != userID {
		return "", fmt.Errorf("asset does not belong to user")
	}

	// 只有文件类型的资产才需要生成预签名URL
	if asset.Source != "file" {
		return "", fmt.Errorf("only file assets can generate presigned URL")
	}

	// 如果没有存储配置ID，无法生成预签名URL
	if asset.StorageConfigID == nil {
		return "", fmt.Errorf("asset has no storage config")
	}

	// 获取OSS客户端
	ossClient, storageConfig, err := s.ossService.GetOSSClientByConfigID(ctx, *asset.StorageConfigID)
	if err != nil {
		return "", fmt.Errorf("failed to get OSS client: %w", err)
	}

	// 从 StorageURL 中提取 objectKey
	objectKey := s.extractObjectKeyFromURL(asset.StorageURL, storageConfig.BaseURL)
	if objectKey == "" {
		return "", fmt.Errorf("failed to extract object key from storage URL")
	}

	// 生成1小时的预签名URL
	presignedURL, err := ossClient.GeneratePresignedURL(ctx, storageConfig.Bucket, objectKey, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL, nil
}

// 辅助方法

// generateFileHash 生成文件哈希
func (s *AssetService) generateFileHash(fileName, userID string, timestamp int64) string {
	data := fmt.Sprintf("%s_%s_%d", fileName, userID, timestamp)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])[:12] // 取前12位
}

// generateAssetID 生成资产ID
func (s *AssetService) generateAssetID(userID string, timestamp int64) string {
	data := fmt.Sprintf("%s_%d_%d", userID, timestamp, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("asset_%s", hex.EncodeToString(hash[:])[:16])
}

// detectAssetType 从文件名和MIME类型检测资产类型
func (s *AssetService) detectAssetType(fileName, contentType string) string {
	ext := filepath.Ext(fileName)
	extLower := ""
	if len(ext) > 0 {
		extLower = ext[1:] // 去掉点号
	}

	// 优先从MIME类型判断
	if contentType != "" {
		if len(contentType) >= 5 && contentType[:5] == "image" {
			return "image"
		}
		if len(contentType) >= 5 && contentType[:5] == "audio" {
			return "audio"
		}
		if len(contentType) >= 5 && contentType[:5] == "video" {
			return "video"
		}
	}

	// 从文件扩展名判断
	imageExts := map[string]bool{
		"jpg": true, "jpeg": true, "png": true, "gif": true,
		"bmp": true, "webp": true, "svg": true,
	}
	audioExts := map[string]bool{
		"mp3": true, "wav": true, "ogg": true, "aac": true,
		"flac": true, "m4a": true,
	}
	videoExts := map[string]bool{
		"mp4": true, "webm": true, "mov": true, "avi": true,
		"flv": true, "mkv": true, "wmv": true,
	}

	if imageExts[extLower] {
		return "image"
	}
	if audioExts[extLower] {
		return "audio"
	}
	if videoExts[extLower] {
		return "video"
	}

	return "other"
}

// detectAssetTypeFromURL 从URL检测资产类型
func (s *AssetService) detectAssetTypeFromURL(url string) string {
	return s.detectAssetType(url, "")
}

// extractObjectKeyFromURL 从存储URL中提取objectKey
func (s *AssetService) extractObjectKeyFromURL(storageURL, baseURL string) string {
	// 如果有 BaseURL，从 storageURL 中移除 baseURL 前缀
	if baseURL != "" {
		// 确保 baseURL 以 / 结尾
		baseURLNormalized := strings.TrimSuffix(baseURL, "/")
		if strings.HasPrefix(storageURL, baseURLNormalized) {
			objectKey := strings.TrimPrefix(storageURL[len(baseURLNormalized):], "/")
			return objectKey
		}
	}

	// 使用标准库解析URL
	parsedURL, err := url.Parse(storageURL)
	if err != nil {
		hlog.Errorf("Failed to parse storage URL: %v", err)
		return ""
	}

	// 提取路径部分（去掉前导斜杠）
	path := strings.TrimPrefix(parsedURL.Path, "/")
	return path
}
