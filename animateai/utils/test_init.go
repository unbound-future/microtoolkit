package utils

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnimateAIPlatform/animate-ai/common/apollo"
	"github.com/AnimateAIPlatform/animate-ai/common/client"
	"github.com/AnimateAIPlatform/animate-ai/common/ctxlogger"
	"github.com/AnimateAIPlatform/animate-ai/common/db"
	"github.com/AnimateAIPlatform/animate-ai/models"

	common_consts "github.com/AnimateAIPlatform/animate-ai/common/consts"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	// testDBInitialized 标记数据库是否已初始化
	testDBInitialized bool
)

// getProjectRoot 获取项目根目录
func getProjectRoot() (string, error) {
	// 通过 go list 获取模块目录
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	output, err := cmd.Output()
	if err != nil {
		// 如果 go list 失败，尝试从当前工作目录向上查找 go.mod
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir := wd
		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		return wd, nil
	}
	return strings.TrimSpace(string(output)), nil
}

// InitTestEnv 初始化测试环境
// 类似 main.go 的 init 函数，但更适合测试场景
func InitTestEnv(t *testing.T) {
	// 初始化日志
	err := ctxlogger.InitDefaultLogger(common_consts.LogFilenPath)
	if err != nil {
		t.Logf("Warning: Failed to initialize logger: %v", err)
		// 测试环境不强制要求日志初始化成功
	}

	// 解析环境配置文件路径（如果提供）
	envConfPath := flag.String("env", "", "env config path")
	flag.Parse()

	// 如果没有指定 env 参数，尝试从项目根目录查找 .env 文件
	if *envConfPath == "" {
		projectRoot, err := getProjectRoot()
		if err == nil {
			envFile := filepath.Join(projectRoot, common_consts.EnvConfFile)
			if _, err := os.Stat(envFile); err == nil {
				*envConfPath = envFile
				hlog.Infof("Found .env file at project root: %s", envFile)
			} else {
				// 尝试当前工作目录
				wd, _ := os.Getwd()
				envFile = filepath.Join(wd, common_consts.EnvConfFile)
				if _, err := os.Stat(envFile); err == nil {
					*envConfPath = envFile
					hlog.Infof("Found .env file at current directory: %s", envFile)
				} else {
					// 使用默认值
					*envConfPath = common_consts.EnvConfFile
				}
			}
		} else {
			*envConfPath = common_consts.EnvConfFile
		}
	}

	if *envConfPath != "" {
		hlog.Infof("Test env config path: %s", *envConfPath)
		common_consts.SetEnvConfFile(*envConfPath)
	}

	// 初始化全局环境变量
	err = common_consts.Init()
	if err != nil {
		t.Logf("Warning: Failed to init global envs: %v", err)
		// 测试环境不强制要求环境变量初始化成功
	}

	// 初始化 HTTP 客户端
	err = client.InitHttpClient()
	if err != nil {
		t.Logf("Warning: Failed to initialize HTTP client: %v", err)
		// 测试环境不强制要求 HTTP 客户端初始化成功
	}

	// 尝试加载数据库配置并初始化（可选）
	if !testDBInitialized {
		var dbConfig models.StaticDBConfigKey
		err = apollo.GetValueFromEnvAndApollo(&dbConfig)
		if err != nil {
			t.Logf("Info: Failed to load static_db_config: %v, database initialization skipped", err)
			return
		}

		// 验证配置完整性
		if dbConfig.Username == "" || dbConfig.Host == "" || dbConfig.Database == "" || dbConfig.Port == 0 {
			t.Logf("Info: static_db_config is incomplete, database initialization skipped")
			return
		}

		// 将 StaticDBConfigKey 转换为 db.Config
		dbCfg := &db.Config{
			User:     dbConfig.Username,
			Password: dbConfig.Password,
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			DBName:   dbConfig.Database,
		}

		err = db.InitDB(dbCfg)
		if err != nil {
			t.Logf("Warning: Failed to initialize database: %v", err)
			return
		}
		t.Logf("Database initialized successfully for testing")

		// 自动创建表
		err = models.InitTables()
		if err != nil {
			t.Logf("Warning: Failed to initialize tables: %v", err)
		} else {
			t.Logf("Tables initialized successfully for testing")
		}

		testDBInitialized = true
	}
}

// MustInitTestEnv 初始化测试环境，失败则跳过测试
// 如果数据库未初始化，会跳过测试
func MustInitTestEnv(t *testing.T) {
	InitTestEnv(t)
	if db.DB == nil {
		t.Skip("Database not initialized, skipping tests that require database")
	}
}
