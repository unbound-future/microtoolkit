package testutil

import (
	f "flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AnimateAIPlatform/animate-ai/common/apollo"
	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/common/ctxlogger"
	"github.com/AnimateAIPlatform/animate-ai/common/format"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func ApolloInit() {
	// 通过go list获取当前项目目录路径
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("执行命令失败: %v", err)
	}
	moduleDir := strings.TrimSpace(string(output))

	// 添加日志在当前路径下logs目录
	err = ctxlogger.InitDefaultLogger(filepath.Join(moduleDir, "logs"))
	if err != nil {
		log.Fatal("Error initializing logger:", err.Error())
		os.Exit(1)
	}

	// env环境变量从当前目录下test.env获取
	envConfPath := f.String("env", filepath.Join(moduleDir, "test.env"), "env config file")

	// 设置 flag 解析错误处理方式，允许忽略未知参数
	f.CommandLine.Init("", f.ContinueOnError)

	f.Parse()
	hlog.Infof("env config path: %s", *envConfPath)

	// 设置环境配置文件路径
	consts.SetEnvConfFile(*envConfPath)

	err = consts.Init()
	if err != nil {
		hlog.Errorf("base env init error: %s", err.Error())
	}

	namespaces := strings.Split(consts.GlobalEnvs.ApolloNamespace, ",")
	_, err = apollo.Init(consts.GlobalEnvs.ApolloAppID, consts.GlobalEnvs.ApolloCluster, namespaces, consts.GlobalEnvs.ApolloMetaAddr)
	if err != nil {
		hlog.Errorf("Warning: Apollo initialization failed: %v", err.Error())
		os.Exit(1)
	}

	configVars := apollo.GetGlobalConfigManager().GetAllVariables()
	configJSON, err := format.FormatMapValueJson(configVars)
	if err != nil {
		hlog.Errorf("Error marshalling apollo config: %s", err.Error())

		os.Exit(1)
	}
	hlog.Infof("apollo initialized: %s", string(configJSON))

}
