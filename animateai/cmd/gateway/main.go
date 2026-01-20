package main

import (
	"flag"
	"log"
	"os"

	"github.com/AnimateAIPlatform/animate-ai/common/apollo"
	"github.com/AnimateAIPlatform/animate-ai/common/client"
	"github.com/AnimateAIPlatform/animate-ai/common/ctxlogger"
	"github.com/AnimateAIPlatform/animate-ai/common/db"
	"github.com/AnimateAIPlatform/animate-ai/common/metrics"
	"github.com/AnimateAIPlatform/animate-ai/internal/gateway"
	"github.com/AnimateAIPlatform/animate-ai/models"

	common_consts "github.com/AnimateAIPlatform/animate-ai/common/consts"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertz_prometheus "github.com/hertz-contrib/monitor-prometheus"
	prometheus "github.com/prometheus/client_golang/prometheus"
)

func init() {

	err := ctxlogger.InitDefaultLogger(common_consts.LogFilenPath)
	if err != nil {
		log.Fatal("Error initializing logger:", err.Error())
		os.Exit(1)
	}

	envConfPath := flag.String("env", common_consts.EnvConfFile, "env config path")
	flag.Parse()
	hlog.Infof("env config path: %s", *envConfPath)

	// 设置环境配置文件路径
	common_consts.SetEnvConfFile(*envConfPath)

	// 初始化全局环境变量（Apollo配置变为可选，不再强制要求）
	err = common_consts.Init()
	if err != nil {
		hlog.Errorf("Failed to init global envs: %v", err)
		os.Exit(1)
	}

	err = client.InitHttpClient()
	if err != nil {
		hlog.Errorf("Error initializing HTTP client: %s", err.Error())
		os.Exit(1)
	}
	hlog.Infof("HTTP client initialized successfully")

	// 加载 static_db_config 配置并初始化 MySQL
	var dbConfig models.StaticDBConfigKey
	err = apollo.GetValueFromEnvAndApollo(&dbConfig)
	if err != nil {
		hlog.Warnf("Failed to load static_db_config: %v, database initialization skipped", err)
	} else {
		// 验证配置完整性
		if dbConfig.Username == "" || dbConfig.Host == "" || dbConfig.Database == "" || dbConfig.Port == 0 {
			hlog.Warnf("static_db_config is incomplete (missing required fields), database initialization skipped")
		} else {
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
				hlog.Errorf("Error initializing database: %s", err.Error())
				os.Exit(1)
			}
			hlog.Infof("Database initialized successfully")

			// 自动创建表
			err = models.InitTables()
			if err != nil {
				hlog.Errorf("Error initializing tables: %s", err.Error())
				os.Exit(1)
			}
			hlog.Infof("Tables initialized successfully")
		}
	}

}

func main() {
	reg := prometheus.NewRegistry()
	metrics.RegisterMetrics(reg)

	h := server.Default(
		server.WithTracer(hertz_prometheus.NewServerTracer(":"+common_consts.GlobalEnvs.PrometheusPort, "/metrics", hertz_prometheus.WithRegistry(reg))),
		server.WithHostPorts(":"+common_consts.GlobalEnvs.ServerPort),
		server.WithStreamBody(true),
	)

	gateway.RegisterGatewayRoutes(h)
	h.Spin()
}
