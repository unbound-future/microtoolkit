package oss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// cosClient 腾讯云COS客户端实现
type cosClient struct {
	client *cos.Client
	config Config
}

// NewCOSClient 创建腾讯云COS客户端
func NewCOSClient(config Config) (Client, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	// 构建COS服务URL
	// 如果提供了BaseURL，优先使用；否则根据bucket和region构建
	var bucketURL *url.URL
	var err error

	if config.BaseURL != "" {
		bucketURL, err = url.Parse(config.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}
		if bucketURL == nil {
			return nil, fmt.Errorf("bucketURL is nil after parsing BaseURL")
		}
	} else {
		// 如果没有提供BaseURL，使用默认格式：https://{bucket}.cos.{region}.myqcloud.com
		if config.Bucket == "" {
			return nil, fmt.Errorf("%w: bucket is required for COS", ErrInvalidConfig)
		}
		if config.Region == "" {
			return nil, fmt.Errorf("%w: region is required for COS", ErrInvalidConfig)
		}
		cosURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Bucket, config.Region)
		bucketURL, err = url.Parse(cosURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse COS URL %s: %w", cosURL, err)
		}
		if bucketURL == nil {
			return nil, fmt.Errorf("bucketURL is nil after parsing COS URL: %s", cosURL)
		}
	}

	// 验证 bucketURL 不为 nil
	if bucketURL == nil {
		return nil, fmt.Errorf("bucketURL is nil, cannot create COS client")
	}

	// 创建COS客户端
	baseURL := &cos.BaseURL{BucketURL: bucketURL}
	if baseURL == nil {
		return nil, fmt.Errorf("failed to create BaseURL struct")
	}
	if baseURL.BucketURL == nil {
		return nil, fmt.Errorf("BaseURL.BucketURL is nil")
	}

	httpClient := &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.AccessKey,
			SecretKey: config.SecretKey,
		},
	}
	if httpClient == nil {
		return nil, fmt.Errorf("failed to create HTTP client")
	}

	client := cos.NewClient(baseURL, httpClient)
	if client == nil {
		return nil, fmt.Errorf("COS client is nil after creation")
	}

	// 验证 client.Object 不为 nil
	if client.Object == nil {
		return nil, fmt.Errorf("COS client.Object is nil after creation")
	}

	return &cosClient{
		client: client,
		config: config,
	}, nil
}

// CreateBucket 创建存储桶（如果不存在）
func (c *cosClient) CreateBucket(ctx context.Context, bucketName string) error {
	// 注意：COS的bucket创建通常需要在控制台完成，这里只做检查
	// 如果bucket不存在，会返回错误
	_, err := c.client.Bucket.Head(ctx)
	if err != nil {
		// 如果bucket不存在，尝试创建
		opts := &cos.BucketPutOptions{
			XCosACL: "private", // 默认私有
		}
		_, err = c.client.Bucket.Put(ctx, opts)
		if err != nil {
			return NewStorageError(StorageTypeCOS, "CreateBucketFailed", fmt.Sprintf("failed to create bucket %s", bucketName), err)
		}
	}
	return nil
}

// Upload 上传文件到对象存储
func (c *cosClient) Upload(ctx context.Context, bucketName, objectKey string, reader io.Reader, contentType string) (string, error) {
	// 安全检查
	if c == nil {
		return "", NewStorageError(StorageTypeCOS, "UploadFailed", "cosClient is nil", nil)
	}
	if c.client == nil {
		return "", NewStorageError(StorageTypeCOS, "UploadFailed", "COS client is not initialized", nil)
	}
	if c.client.Object == nil {
		return "", NewStorageError(StorageTypeCOS, "UploadFailed", "COS Object service is not initialized", nil)
	}
	if objectKey == "" {
		return "", NewStorageError(StorageTypeCOS, "UploadFailed", "objectKey is empty", nil)
	}
	if reader == nil {
		return "", NewStorageError(StorageTypeCOS, "UploadFailed", "reader is nil", nil)
	}

	// 构建上传选项
	var opts *cos.ObjectPutOptions
	if contentType != "" {
		opts = &cos.ObjectPutOptions{
			ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
				ContentType: contentType,
			},
		}
	}

	// 上传文件（opts 可以为 nil，如果为空则使用默认选项）
	_, err := c.client.Object.Put(ctx, objectKey, reader, opts)
	if err != nil {
		return "", NewStorageError(StorageTypeCOS, "UploadFailed", fmt.Sprintf("failed to upload object %s", objectKey), err)
	}

	// 上传成功后，生成对象URL
	// 注意：在 GetObjectURL 之前确保 c 和 c.config 不为 nil
	if c == nil {
		return "", NewStorageError(StorageTypeCOS, "UploadFailed", "cosClient is nil in GetObjectURL", nil)
	}

	storageURL := c.GetObjectURL(ctx, bucketName, objectKey)
	if storageURL == "" {
		return "", NewStorageError(StorageTypeCOS, "UploadFailed", "failed to generate storage URL", nil)
	}
	return storageURL, nil
}

// UploadFromFile 从本地文件上传到对象存储
func (c *cosClient) UploadFromFile(ctx context.Context, bucketName, objectKey, filePath, contentType string) (string, error) {
	// 构建上传选项
	var opts *cos.ObjectPutOptions
	if contentType != "" {
		opts = &cos.ObjectPutOptions{
			ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
				ContentType: contentType,
			},
		}
	}

	// 上传文件（opts 可以为 nil，如果为空则使用默认选项）
	_, err := c.client.Object.PutFromFile(ctx, objectKey, filePath, opts)
	if err != nil {
		return "", NewStorageError(StorageTypeCOS, "UploadFromFileFailed", fmt.Sprintf("failed to upload file %s to %s", filePath, objectKey), err)
	}

	// 返回对象URL
	return c.GetObjectURL(ctx, bucketName, objectKey), nil
}

// GeneratePresignedURL 生成预签名下载链接
func (c *cosClient) GeneratePresignedURL(ctx context.Context, bucketName, objectKey string, expiration time.Duration) (string, error) {
	// 生成预签名URL
	presignedURL, err := c.client.Object.GetPresignedURL(ctx, http.MethodGet, objectKey, c.config.AccessKey, c.config.SecretKey, expiration, nil)
	if err != nil {
		return "", NewStorageError(StorageTypeCOS, "GeneratePresignedURLFailed", fmt.Sprintf("failed to generate presigned URL for %s", objectKey), err)
	}

	return presignedURL.String(), nil
}

// Delete 删除对象
func (c *cosClient) Delete(ctx context.Context, bucketName, objectKey string) error {
	_, err := c.client.Object.Delete(ctx, objectKey)
	if err != nil {
		return NewStorageError(StorageTypeCOS, "DeleteFailed", fmt.Sprintf("failed to delete object %s", objectKey), err)
	}
	return nil
}

// Exists 检查对象是否存在
func (c *cosClient) Exists(ctx context.Context, bucketName, objectKey string) (bool, error) {
	exists, err := c.client.Object.IsExist(ctx, objectKey)
	if err != nil {
		return false, NewStorageError(StorageTypeCOS, "ExistsCheckFailed", fmt.Sprintf("failed to check existence of object %s", objectKey), err)
	}
	return exists, nil
}

// GetObjectURL 获取对象的公开访问URL（如果对象是公开的）
func (c *cosClient) GetObjectURL(ctx context.Context, bucketName, objectKey string) string {
	// 安全检查：确保 config 不为 nil（虽然不应该发生）
	if c == nil {
		return ""
	}

	// 优先使用传入的bucketName，如果没有则使用配置中的bucket
	actualBucket := bucketName
	if actualBucket == "" {
		actualBucket = c.config.Bucket
	}

	// 如果 actualBucket 仍然为空，返回空字符串或错误格式
	if actualBucket == "" {
		return ""
	}

	// 如果配置了BaseURL，使用BaseURL
	if c.config.BaseURL != "" && len(c.config.BaseURL) > 0 {
		baseURL := c.config.BaseURL
		// 确保BaseURL以/结尾
		if len(baseURL) > 0 && baseURL[len(baseURL)-1] != '/' {
			baseURL += "/"
		}
		return baseURL + objectKey
	}

	// 否则使用默认格式
	region := c.config.Region
	if region == "" {
		region = "ap-beijing" // 默认区域
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", actualBucket, region, objectKey)
}

// ListObjects 列出存储桶中的对象
func (c *cosClient) ListObjects(ctx context.Context, bucketName, prefix, marker string, maxKeys int) ([]ObjectInfo, error) {
	// 构建列表选项
	opts := &cos.BucketGetOptions{
		Prefix:  prefix,
		Marker:  marker,
		MaxKeys: maxKeys,
	}

	// 如果 maxKeys 为 0，使用默认值
	if maxKeys <= 0 {
		opts.MaxKeys = 1000
	}

	// 列出对象
	result, _, err := c.client.Bucket.Get(ctx, opts)
	if err != nil {
		return nil, NewStorageError(StorageTypeCOS, "ListObjectsFailed", fmt.Sprintf("failed to list objects in bucket %s", bucketName), err)
	}

	// 转换结果
	objects := make([]ObjectInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		// 解析 LastModified 时间（COS SDK 返回的是 RFC3339 格式的字符串）
		var lastModified time.Time
		if obj.LastModified != "" {
			parsedTime, err := time.Parse(time.RFC3339, obj.LastModified)
			if err != nil {
				// 如果解析失败，尝试其他常见格式
				parsedTime, err = time.Parse("2006-01-02T15:04:05.000Z", obj.LastModified)
				if err != nil {
					// 解析失败时使用零值
					lastModified = time.Time{}
				} else {
					lastModified = parsedTime
				}
			} else {
				lastModified = parsedTime
			}
		}

		objects = append(objects, ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: lastModified,
			ETag:         obj.ETag,
			StorageClass: obj.StorageClass,
		})
	}

	return objects, nil
}
