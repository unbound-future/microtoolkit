package oss

import "time"

// StorageType 存储类型常量
const (
	StorageTypeOSS   = "oss"   // 阿里云OSS
	StorageTypeS3    = "s3"    // AWS S3
	StorageTypeCOS   = "cos"   // 腾讯云COS
	StorageTypeMinio = "minio" // MinIO
)

// UploadOptions 上传选项
type UploadOptions struct {
	ContentType        string            // 文件MIME类型
	Metadata           map[string]string // 对象元数据
	ACL                string            // 访问控制列表（如 "public-read", "private" 等）
	CacheControl       string            // 缓存控制
	ContentDisposition string            // 内容处置（如 "attachment; filename=xxx"）
}

// PresignedURLOptions 预签名URL选项
type PresignedURLOptions struct {
	Method      string            // HTTP方法（GET, PUT, POST等），默认为GET
	Expiration  time.Duration     // 过期时间
	Headers     map[string]string // 额外的请求头
	QueryParams map[string]string // 额外的查询参数
}

