package gateway

import (
	"github.com/AnimateAIPlatform/animate-ai/internal/gateway/handler"
	"github.com/AnimateAIPlatform/animate-ai/middleware/auth"
	"github.com/AnimateAIPlatform/animate-ai/middleware/cors"
	"github.com/AnimateAIPlatform/animate-ai/middleware/logger"
	"github.com/AnimateAIPlatform/animate-ai/middleware/retry"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// RegisterGatewayRoutes registers all gateway routes
func RegisterGatewayRoutes(h *server.Hertz) {
	// CORS 中间件（必须在最前面，应用到所有路由和所有 HTTP 方法）
	// 这会确保所有接口（包括未定义的路由）都经过 CORS 处理
	h.Use(cors.CORS())

	// Health check
	h.Use(logger.AccessLog())
	h.Use(retry.Retry())

	// User routes
	api := h.Group("/api")
	user := api.Group("/user")
	
	// 注册接口不需要鉴权
	user.POST("/register", handler.Register)
	
	// 登录接口不需要鉴权（因为鉴权需要先登录）
	user.POST("/login", handler.Login)
	
	// 其他接口都需要鉴权
	user.Use(auth.Auth())
	user.GET("/userInfo", handler.GetUserInfo)
	user.POST("/saveInfo", handler.SaveUserInfo)
	user.GET("/list", handler.ListUsers)
	user.DELETE("/:userId", handler.DeleteUser)
	user.POST("/delete", handler.DeleteUser) // 也支持 POST 方式删除

	// Asset routes 资产管理路由
	asset := api.Group("/asset")
	asset.Use(auth.Auth()) // 所有资产接口都需要鉴权
	asset.POST("/upload", handler.UploadAsset)        // 上传文件资产
	asset.POST("/add-by-url", handler.AddAssetByURL)  // 通过URL添加资产
	asset.GET("/list", handler.ListAssets)                          // 列出用户资产
	asset.GET("/:assetId/presigned-url", handler.GeneratePresignedURL) // 生成预签名下载链接（必须在/:assetId之前）
	asset.GET("/:assetId", handler.GetAsset)                        // 获取资产详情
	asset.PUT("/:assetId", handler.UpdateAsset)                     // 更新资产信息
	asset.DELETE("/:assetId", handler.DeleteAsset)                  // 删除资产

	// Tool Component routes 工具组件路由
	toolComponent := api.Group("/tool-component")
	toolComponent.Use(auth.Auth()) // 所有工具组件接口都需要鉴权
	toolComponent.POST("", handler.CreateToolComponent)           // 创建工具组件
	toolComponent.GET("/list", handler.ListToolComponents)        // 列出用户工具组件
	toolComponent.GET("/:componentId", handler.GetToolComponent)  // 获取工具组件详情
	toolComponent.PUT("/:componentId", handler.UpdateToolComponent) // 更新工具组件信息
	toolComponent.DELETE("/:componentId", handler.DeleteToolComponent) // 删除工具组件

	// Agent Flow routes 工作流路由
	agentFlow := api.Group("/agent-flow")
	agentFlow.Use(auth.Auth()) // 所有工作流接口都需要鉴权
	agentFlow.POST("", handler.CreateAgentFlow)           // 创建工作流
	agentFlow.GET("/list", handler.ListAgentFlows)        // 列出用户工作流
	agentFlow.GET("/:flowId", handler.GetAgentFlow)       // 获取工作流详情
	agentFlow.PUT("/:flowId", handler.UpdateAgentFlow)    // 更新工作流信息
	agentFlow.DELETE("/:flowId", handler.DeleteAgentFlow) // 删除工作流

	// Workflow Template routes 工作流模版路由
	workflowTemplate := api.Group("/workflow-template")
	workflowTemplate.Use(auth.Auth()) // 所有工作流模版接口都需要鉴权
	workflowTemplate.POST("", handler.CreateWorkflowTemplate)           // 创建工作流模版
	workflowTemplate.GET("/list", handler.ListWorkflowTemplates)        // 列出用户工作流模版
	workflowTemplate.GET("/:templateId", handler.GetWorkflowTemplate)   // 获取工作流模版详情
	workflowTemplate.PUT("/:templateId", handler.UpdateWorkflowTemplate) // 更新工作流模版信息
	workflowTemplate.DELETE("/:templateId", handler.DeleteWorkflowTemplate) // 删除工作流模版
}
