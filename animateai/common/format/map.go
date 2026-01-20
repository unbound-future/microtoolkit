package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
)

// FormatApolloConfig 格式化Apollo配置，避免JSON字符串被双重转义
func FormatMapValueJson(m map[string]string) ([]byte, error) {
	// 创建格式化后的配置映射，避免JSON字符串被双重转义
	formattedConfig := make(map[string]interface{})
	for key, value := range m {
		// 尝试解析为JSON，如果成功则使用解析后的对象，否则使用原字符串
		var jsonData interface{}
		if err := json.Unmarshal([]byte(value), &jsonData); err == nil {
			formattedConfig[key] = jsonData
		} else {
			formattedConfig[key] = value
		}
	}

	// 序列化格式化后的配置，避免转义
	return sonic.Marshal(&formattedConfig)
}

func TruncatedBody(body string) string { //, contentType string) string {
	// if strings.Contains(contentType, "multipart/form-data") {
	// 	return ParseMultipartFormData([]byte(body), contentType)
	// } else
	//  {
	// 尝试解析为JSON
	var bodyData interface{}
	if err := json.Unmarshal([]byte(body), &bodyData); err == nil {
		// 对JSON数据使用ProcessMapValues处理
		processedData := ProcessMapValues(bodyData)
		return FormatValue(processedData)
	} else {
		// 对非JSON内容，使用新的截断逻辑
		return truncateNonJsonBody(body)
	}
	// }
}

func FormatValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	// 使用标准JSON格式输出，并去掉换行符
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	// 去掉所有类型的换行符，让日志输出更紧凑
	result := strings.ReplaceAll(string(bytes), "\n", "")
	result = strings.ReplaceAll(result, "\r\n", "")
	result = strings.ReplaceAll(result, "\r", "")
	return result
}

// truncateNonJsonBody 截断非JSON内容，保留首尾各100字符
func truncateNonJsonBody(bodyStr string) string {
	const headSize = 300
	const tailSize = 100
	const minTruncateSize = headSize + tailSize

	if len(bodyStr) <= minTruncateSize {
		return bodyStr
	}

	// 直接截取首尾各100字符
	headPart := bodyStr[:headSize]
	tailPart := bodyStr[len(bodyStr)-tailSize:]

	// 使用空格分隔，避免引入换行符
	result := fmt.Sprintf("%s ...[truncated, total: %d chars, showing first %d and last %d chars]... %s",
		headPart, len(bodyStr), headSize, tailSize, tailPart)

	// 去掉所有类型的换行符，让日志输出更紧凑
	result = strings.ReplaceAll(result, "\n", " ")
	result = strings.ReplaceAll(result, "\r\n", " ")
	result = strings.ReplaceAll(result, "\r", " ")
	return result
}

// processMapValues 递归处理map中的所有值，截断长字符串
func ProcessMapValues(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = ProcessMapValues(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = ProcessMapValues(value)
		}
		return result
	case string:
		if len(v) > 100 {
			return v[:100] + fmt.Sprintf("...[truncated, total: %d chars]", len(v))
		}
		return v
	default:
		return v
	}
}

// parseMultipartFormData 解析 multipart/form-data 请求体
func ParseMultipartFormData(body []byte, contentType string) string {
	// 对于 multipart/form-data，我们只显示基本信息，避免解析消耗数据
	boundary := getBoundary(contentType)
	if boundary == "" {
		return fmt.Sprintf("[multipart/form-data - no boundary, body size: %d bytes]", len(body))
	}

	// 简单统计字段数量，不进行详细解析
	bodyStr := string(body)
	fieldCount := strings.Count(bodyStr, "--"+boundary) - 1 // 减去最后的结束边界

	if fieldCount <= 0 {
		return fmt.Sprintf("[multipart/form-data - no fields, body size: %d bytes]", len(body))
	}

	return fmt.Sprintf("[multipart/form-data - %d fields, body size: %d bytes]", fieldCount, len(body))
}

// getBoundary 从 Content-Type 中提取 boundary
func getBoundary(contentType string) string {
	parts := strings.Split(contentType, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "boundary=") {
			return strings.TrimPrefix(part, "boundary=")
		}
	}
	return ""
}
