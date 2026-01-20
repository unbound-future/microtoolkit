package oss

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrUnsupportedStorageType 不支持的存储类型错误
	ErrUnsupportedStorageType = errors.New("unsupported storage type")
	// ErrInvalidConfig 无效的配置错误
	ErrInvalidConfig = errors.New("invalid storage config")
	// ErrBucketNotFound 存储桶不存在错误
	ErrBucketNotFound = errors.New("bucket not found")
	// ErrObjectNotFound 对象不存在错误
	ErrObjectNotFound = errors.New("object not found")
	// ErrUploadFailed 上传失败错误
	ErrUploadFailed = errors.New("upload failed")
	// ErrDeleteFailed 删除失败错误
	ErrDeleteFailed = errors.New("delete failed")
)

// StorageError 存储操作错误
type StorageError struct {
	Type    string // 错误类型：oss, s3, cos, minio 等
	Code    string // 错误代码
	Message string // 错误消息
	Err     error  // 原始错误
}

func (e *StorageError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %s (original: %v)", e.Type, e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

// NewStorageError 创建存储错误
func NewStorageError(storageType, code, message string, err error) *StorageError {
	return &StorageError{
		Type:    storageType,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ValidateConfig 验证配置是否有效
func ValidateConfig(config Config) error {
	if config.Type == "" {
		return fmt.Errorf("%w: storage type is required", ErrInvalidConfig)
	}
	// 对于COS，可以使用BaseURL或Region+Bucket，不需要Endpoint
	if config.Type != StorageTypeCOS && config.Endpoint == "" && config.BaseURL == "" {
		return fmt.Errorf("%w: endpoint or baseURL is required for %s", ErrInvalidConfig, config.Type)
	}
	if config.AccessKey == "" {
		return fmt.Errorf("%w: access key is required", ErrInvalidConfig)
	}
	if config.SecretKey == "" {
		return fmt.Errorf("%w: secret key is required", ErrInvalidConfig)
	}
	return nil
}

// NewClientFromModel 从数据库模型创建客户端
func NewClientFromModel(ctx context.Context, config interface{}) (Client, error) {
	// 这里可以根据实际的模型结构进行转换
	// 示例：如果传入的是 models.ObjectStorageConfig
	// 需要根据实际情况实现
	return nil, fmt.Errorf("not implemented: use NewClient instead")
}
