package oss

import (
	"context"
	"fmt"
	"io"
	"time"
)

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key          string    // 对象键名
	Size         int64     // 对象大小（字节）
	LastModified time.Time // 最后修改时间
	ETag         string    // ETag
	StorageClass string    // 存储类型
}

// Client 对象存储客户端接口
// 支持多家云厂商的统一接口抽象
type Client interface {
	// CreateBucket 创建存储桶（如果不存在）
	// ctx: 上下文
	// bucketName: 存储桶名称
	// 返回错误如果创建失败
	CreateBucket(ctx context.Context, bucketName string) error

	// Upload 上传文件到对象存储
	// ctx: 上下文
	// bucketName: 存储桶名称
	// objectKey: 对象键名（文件路径）
	// reader: 文件内容读取器
	// contentType: 文件MIME类型（可选，如 "image/jpeg", "video/mp4" 等）
	// 返回上传后的对象URL和错误
	Upload(ctx context.Context, bucketName, objectKey string, reader io.Reader, contentType string) (string, error)

	// UploadFromFile 从本地文件上传到对象存储
	// ctx: 上下文
	// bucketName: 存储桶名称
	// objectKey: 对象键名（文件路径）
	// filePath: 本地文件路径
	// contentType: 文件MIME类型（可选）
	// 返回上传后的对象URL和错误
	UploadFromFile(ctx context.Context, bucketName, objectKey, filePath, contentType string) (string, error)

	// GeneratePresignedURL 生成预签名下载链接
	// ctx: 上下文
	// bucketName: 存储桶名称
	// objectKey: 对象键名（文件路径）
	// expiration: 链接过期时间（从当前时间开始计算）
	// 返回预签名URL和错误
	GeneratePresignedURL(ctx context.Context, bucketName, objectKey string, expiration time.Duration) (string, error)

	// Delete 删除对象
	// ctx: 上下文
	// bucketName: 存储桶名称
	// objectKey: 对象键名（文件路径）
	// 返回错误如果删除失败
	Delete(ctx context.Context, bucketName, objectKey string) error

	// Exists 检查对象是否存在
	// ctx: 上下文
	// bucketName: 存储桶名称
	// objectKey: 对象键名（文件路径）
	// 返回是否存在和错误
	Exists(ctx context.Context, bucketName, objectKey string) (bool, error)

	// GetObjectURL 获取对象的公开访问URL（如果对象是公开的）
	// ctx: 上下文
	// bucketName: 存储桶名称
	// objectKey: 对象键名（文件路径）
	// 返回对象URL
	GetObjectURL(ctx context.Context, bucketName, objectKey string) string

	// ListObjects 列出存储桶中的对象
	// ctx: 上下文
	// bucketName: 存储桶名称
	// prefix: 对象键前缀（可选，用于过滤）
	// marker: 分页标记（可选，用于分页）
	// maxKeys: 最大返回数量（可选，默认1000）
	// 返回对象列表和错误
	ListObjects(ctx context.Context, bucketName, prefix, marker string, maxKeys int) ([]ObjectInfo, error)
}

// Config 对象存储配置
type Config struct {
	Type      string // 存储类型：oss, s3, cos, minio 等
	Endpoint  string // 存储服务端点
	Bucket    string // 存储桶名称
	AccessKey string // 访问密钥ID
	SecretKey string // 访问密钥
	Region    string // 区域（可选）
	BaseURL   string // 基础URL（可选，用于自定义域名）
}

// NewClient 根据配置创建对应的对象存储客户端
// 支持多种云厂商：oss, s3, cos, minio 等
// 注意：具体的客户端实现需要在对应的实现文件中提供（如 oss_client.go, s3_client.go 等）
func NewClient(config Config) (Client, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	switch config.Type {
	case StorageTypeOSS:
		return NewOSSClient(config)
	case StorageTypeS3:
		return NewS3Client(config)
	case StorageTypeCOS:
		return NewCOSClient(config)
	case StorageTypeMinio:
		return NewMinioClient(config)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedStorageType, config.Type)
	}
}
