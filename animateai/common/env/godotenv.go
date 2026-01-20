package env

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/joho/godotenv"
)

func LoadEnv(envConfFile string, msutEnv []string) error {
	var err error
	if envConfFile != "" {
		err = godotenv.Load(envConfFile)
		if err != nil {
			hlog.CtxErrorf(context.Background(), "load env file error: %v", err)
		}
	}

	err = nil

	for _, env := range msutEnv {
		if os.Getenv(env) == "" {
			return fmt.Errorf("environment variable %s is not set", env)
		}
	}

	return nil
}

func GetEnvWithDefault(key string, defaultValue string) string {
	env := os.Getenv(key)
	if env == "" {
		return defaultValue
	}
	return env
}

// GetEnvIntWithDefault 获取整型环境变量，如果不存在则返回默认值
func GetEnvIntWithDefault(key string, defaultValue int) int {
	env := os.Getenv(key)
	if env == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(env)
	if err != nil {
		hlog.CtxWarnf(context.Background(), "Invalid integer value for %s: %s, using default: %d", key, env, defaultValue)
		return defaultValue
	}
	return val
}

// GetEnvInt64WithDefault 获取 int64 类型环境变量，如果不存在则返回默认值
func GetEnvInt64WithDefault(key string, defaultValue int64) int64 {
	env := os.Getenv(key)
	if env == "" {
		return defaultValue
	}
	val, err := strconv.ParseInt(env, 10, 64)
	if err != nil {
		hlog.CtxWarnf(context.Background(), "Invalid int64 value for %s: %s, using default: %d", key, env, defaultValue)
		return defaultValue
	}
	return val
}

// GetEnvBoolWithDefault 获取布尔类型环境变量，如果不存在则返回默认值
func GetEnvBoolWithDefault(key string, defaultValue bool) bool {
	env := os.Getenv(key)
	if env == "" {
		return defaultValue
	}
	val, err := strconv.ParseBool(env)
	if err != nil {
		hlog.CtxWarnf(context.Background(), "Invalid boolean value for %s: %s, using default: %v", key, env, defaultValue)
		return defaultValue
	}
	return val
}
