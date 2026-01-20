package consts

import (
	hertzzap "github.com/hertz-contrib/logger/zap"
)

type CtxKey = hertzzap.ExtraKey

const (
	EnvConfFile   = ".env"
	LogFilenPath  = "./logs/"
	LogMaxSize    = 1 * 1024
	LogMaxBackups = 15
	LogMaxAge     = 10
	LogCompress   = true
)

const (
	BillAPPID = "github.com/AnimateAIPlatform/animate-ai.billing"
)

const (
	ClientTraceIDKey    CtxKey = "X-Client-Trace-Id"
	ServerTraceIDKey    CtxKey = "X-Server-Trace-Id"
	ClientRetryCountKey CtxKey = "X-Client-Retry-Count"
	ServerRetryCountKey CtxKey = "X-Server-Retry-Count"
	UpstreamResponseKey CtxKey = "UpstreamResponse"
	UpstreamHeadersKey  CtxKey = "UpstreamHeaders"
	// 用户信息相关 context key
	UserIDKey      CtxKey = "UserID"      // 用户自增ID
	UserAccountIDKey CtxKey = "UserAccountID" // 用户账户ID
	UserNameKey    CtxKey = "UserName"    // 用户名
)

const (
	MaxRetry      = "max-retry"
	RetryInterval = "retry-interval-time-seconds"
)
const (
	HttpClientConfigKey    = "dynamic_http_client_config"
	CacheConfigKey         = "dynamic_cache_config"
	RetryRuleConfigKey     = "dynamic_retry_rule_config"
	DynamicErrorLogMapping = "dynamic_errorlog_mapping"
)
