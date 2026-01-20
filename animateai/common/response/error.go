package response

import (
	"encoding/json"
	"fmt"
)

// ErrorType 错误类型枚举
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation" // 参数验证错误
	ErrorTypeBusiness   ErrorType = "business"   // 业务逻辑错误
	ErrorTypeSystem     ErrorType = "system"     // 系统错误
	ErrorTypeExternal   ErrorType = "external"   // 外部服务错误
	ErrorTypeDatabase   ErrorType = "database"   // 数据库错误
	ErrorTypeNetwork    ErrorType = "network"    // 网络错误
)

// ErrorCode 错误代码枚举
type ErrorCode string

const (
	// 通用错误代码
	ErrorCodeInvalidParam  ErrorCode = "INVALID_PARAM"
	ErrorCodeMissingParam  ErrorCode = "MISSING_PARAM"
	ErrorCodeInternalError ErrorCode = "INTERNAL_ERROR"
	ErrorCodeExternalError ErrorCode = "EXTERNAL_ERROR"
	ErrorCodeDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrorCodeNetworkError  ErrorCode = "NETWORK_ERROR"
)

// APIError 通用API错误结构
type APIError struct {
	ErrCode    ErrorCode `json:"err_code"`
	ErrType    ErrorType `json:"err_type"`
	ErrMessage string    `json:"err_message"`
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.ErrType, e.ErrCode, e.ErrMessage)
}

// ToJSON 将错误转换为JSON格式
func (e *APIError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// NewAPIError 创建新的API错误
func NewAPIError(errCode ErrorCode, errType ErrorType, errMessage string) *APIError {
	return &APIError{
		ErrCode:    errCode,
		ErrType:    errType,
		ErrMessage: errMessage,
	}
}

// NewValidationError 创建参数验证错误
func NewValidationError(errCode ErrorCode, errMessage string) *APIError {
	return NewAPIError(errCode, ErrorTypeValidation, errMessage)
}

// NewBusinessError 创建业务逻辑错误
func NewBusinessError(errCode ErrorCode, errMessage string) *APIError {
	return NewAPIError(errCode, ErrorTypeBusiness, errMessage)
}

// NewSystemError 创建系统错误
func NewSystemError(errCode ErrorCode, errMessage string) *APIError {
	return NewAPIError(errCode, ErrorTypeSystem, errMessage)
}

// NewExternalError 创建外部服务错误
func NewExternalError(errCode ErrorCode, errMessage string) *APIError {
	return NewAPIError(errCode, ErrorTypeExternal, errMessage)
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(errCode ErrorCode, errMessage string) *APIError {
	return NewAPIError(errCode, ErrorTypeDatabase, errMessage)
}

// NewNetworkError 创建网络错误
func NewNetworkError(errCode ErrorCode, errMessage string) *APIError {
	return NewAPIError(errCode, ErrorTypeNetwork, errMessage)
}
