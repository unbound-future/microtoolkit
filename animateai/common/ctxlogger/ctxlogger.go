package ctxlogger

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/common/apollo"
	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/common/format"
	"github.com/AnimateAIPlatform/animate-ai/common/metrics"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertzzap "github.com/hertz-contrib/logger/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitDefaultLogger(logFilePath string) error {
	if err := os.MkdirAll(logFilePath, 0o777); err != nil {
		return err
	}

	// 将文件名设置为日期
	logFileName := time.Now().Format("2006-01-02") + ".log"
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			return err
		}
	}

	// 提供压缩和删除（文件输出）
	lumberjackLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    consts.LogMaxSize,    // 一个文件最大可达 20M。
		MaxBackups: consts.LogMaxBackups, // 最多同时保存 5 个文件。
		MaxAge:     consts.LogMaxAge,     // 一个文件最多可以保存 10 天。
		Compress:   consts.LogCompress,   // 用 gzip 压缩。
	}

	// 基础 Encoder 配置
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncodeCaller = zapcore.FullCallerEncoder

	// 包装为自定义 ctxEncoder（使用控制台格式而非 JSON）
	baseEnc := zapcore.NewConsoleEncoder(cfg)
	// 文件输出的 encoder 负责指标计数
	fileEnc := &ctxEncoder{Encoder: baseEnc, enableMetrics: true}
	// 标准输出的 encoder 不负责指标计数
	stdoutEnc := &ctxEncoder{Encoder: baseEnc, enableMetrics: false}

	// 同时输出到文件和标准输出，便于测试时在控制台看到日志
	logger := hertzzap.NewLogger(
		hertzzap.WithCores(
			hertzzap.CoreConfig{Enc: fileEnc, Ws: zapcore.AddSync(lumberjackLogger), Lvl: zap.NewAtomicLevelAt(zap.DebugLevel)},
			hertzzap.CoreConfig{Enc: stdoutEnc, Ws: zapcore.AddSync(os.Stdout), Lvl: zap.NewAtomicLevelAt(zap.DebugLevel)},
		),
		hertzzap.WithExtraKeys([]hertzzap.ExtraKey{hertzzap.ExtraKey(consts.ServerTraceIDKey)}),
		hertzzap.WithZapOptions(zap.AddCaller(), zap.AddCallerSkip(3)),
	)

	hlog.SetLogger(logger)
	return nil
}

// 自定义 Encoder，在日志中注入 traceID
type ctxEncoder struct {
	zapcore.Encoder
	traceValue    string
	enableMetrics bool // 是否启用指标计数
}

func (e *ctxEncoder) Clone() zapcore.Encoder {
	return &ctxEncoder{Encoder: e.Encoder.Clone(), traceValue: e.traceValue, enableMetrics: e.enableMetrics}
}

// EncodeEntry 自动从 ctx 获取 traceID 并拼接到日志消息
func (e *ctxEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// 查找 traceID 字段，并从输出字段中过滤掉它，避免被额外打印
	var traceID string
	if e.traceValue != "" {
		traceID = string(e.traceValue)
	}
	filtered := fields[:0]
	for _, field := range fields {

		if field.Key == string(consts.ServerTraceIDKey) {
			// 优先从 StringType 读取
			if field.Type == zapcore.StringType && field.String != "" {
				traceID = field.String
				continue // 过滤掉该字段
			}
			// 兜底从 Interface 读取
			if v, ok := field.Interface.(string); ok && v != "" {
				traceID = v
				continue // 过滤掉该字段
			}
			// 未取到有效值则仍然过滤该字段，避免结构化输出
			continue
		}
		filtered = append(filtered, field)
	}

	// 检测错误级别日志并进行错误类型匹配（只在启用指标的 encoder 中进行）
	if e.enableMetrics && ent.Level >= zapcore.ErrorLevel {
		checkAndIncrementErrorMetrics(ent.Message)
	}

	// ERROR 级别日志不截断，其他级别使用 TruncatedBody 处理日志消息
	var finalMessage string
	if ent.Level >= zapcore.ErrorLevel {
		// ERROR 级别日志不截断，保留完整内容
		finalMessage = ent.Message
	} else {
		// 其他级别日志使用 TruncatedBody 处理，截断过长内容
		finalMessage = format.TruncatedBody(ent.Message)
	}

	// 拼接 traceID 到 Message
	if traceID == "" {
		traceID = "NOTraceID"
	}
	ent.Message = fmt.Sprintf("%s	%s", traceID, finalMessage)

	return e.Encoder.EncodeEntry(ent, filtered)
}

// checkAndIncrementErrorMetrics 检查错误日志消息并根据配置增加相应的指标计数
func checkAndIncrementErrorMetrics(message string) {
	// 从 Apollo 配置中心获取 dynamic_errorlog_mapping 配置
	configManager := apollo.GetGlobalConfigManager()
	mappingConfig, exists := configManager.GetVariableWithNamespace(apollo.ApolloNamespaceApplication, consts.DynamicErrorLogMapping)

	var errorMapping map[string]string
	var matched bool

	if exists {
		// 解析 JSON 配置
		if err := json.Unmarshal([]byte(mappingConfig), &errorMapping); err == nil {
			// 遍历错误映射配置，检查消息是否匹配任何正则表达式
			for errorType, regexPattern := range errorMapping {
				regex, err := regexp.Compile(regexPattern)
				if err != nil {
					// 如果正则表达式编译失败，跳过该规则
					continue
				}

				if regex.MatchString(message) {
					// 如果匹配成功，增加对应的 Prometheus 指标计数
					metrics.IncrementErrLogTotalCounter(errorType, 1)
					matched = true
					// 匹配到第一个规则后即停止，避免重复计数
					break
				}
			}
		}
	}

	// 如果没有匹配到任何规则，则使用 "unknown" 作为错误类型
	if !matched {
		metrics.IncrementErrLogTotalCounter("unknown", 1)
	}
}

// 拦截字符串字段的写入，捕获 traceID 并阻止其作为字段输出
func (e *ctxEncoder) AddString(key, value string) {
	if key == string(consts.ServerTraceIDKey) {
		e.traceValue = string(value)
		return
	}
	e.Encoder.AddString(key, value)
}
