package oss

import "fmt"

// 以下函数需要在对应的实现文件中提供具体实现
// 例如：oss_client.go, s3_client.go, cos_client.go, minio_client.go

// NewOSSClient 创建阿里云OSS客户端
// 实现文件：oss_client.go
func NewOSSClient(config Config) (Client, error) {
	return nil, fmt.Errorf("OSS client not implemented yet, please implement in oss_client.go")
}

// NewS3Client 创建AWS S3客户端
// 实现文件：s3_client.go
func NewS3Client(config Config) (Client, error) {
	return nil, fmt.Errorf("S3 client not implemented yet, please implement in s3_client.go")
}

// NewCOSClient 创建腾讯云COS客户端
// 实现文件：cos_client.go
// 注意：此函数已在 cos_client.go 中实现

// NewMinioClient 创建MinIO客户端
// 实现文件：minio_client.go
func NewMinioClient(config Config) (Client, error) {
	return nil, fmt.Errorf("MinIO client not implemented yet, please implement in minio_client.go")
}
