package handler

import (
	"context"
	"fmt"

	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/internal/gateway/service"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hzconsts "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// CreateToolComponentRequest 创建工具组件请求
type CreateToolComponentRequest struct {
	Name           string `json:"name" binding:"required"`   // 组件名称
	Description    string `json:"description"`               // 组件描述（可选）
	Type           string `json:"type" binding:"required"`   // 组件类型：asset、service 或 trigger
	AssetID        string `json:"asset_id,omitempty"`        // 资产ID（资产组件类型时使用）
	ServiceURL     string `json:"service_url,omitempty"`     // 服务URL（服务组件类型时使用）
	ParamDesc      string `json:"param_desc,omitempty"`      // 参数说明（服务组件类型时使用）
	CronExpression string `json:"cron_expression,omitempty"` // Cron表达式（时间触发器类型时使用）
}

// UpdateToolComponentRequest 更新工具组件请求
type UpdateToolComponentRequest struct {
	Name           string `json:"name" binding:"required"`   // 组件名称
	Description    string `json:"description"`               // 组件描述（可选）
	AssetID        string `json:"asset_id,omitempty"`        // 资产ID（资产组件类型时使用）
	ServiceURL     string `json:"service_url,omitempty"`     // 服务URL（服务组件类型时使用）
	ParamDesc      string `json:"param_desc,omitempty"`      // 参数说明（服务组件类型时使用）
	CronExpression string `json:"cron_expression,omitempty"` // Cron表达式（时间触发器类型时使用）
}

// ToolComponentResponse 工具组件响应
type ToolComponentResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Msg    string      `json:"msg,omitempty"`
}

// CreateToolComponent 创建工具组件接口
// POST /api/tool-component
func CreateToolComponent(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, ToolComponentResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	var req CreateToolComponentRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid request parameters: %v", err)
		c.JSON(hzconsts.StatusBadRequest, ToolComponentResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	componentService := service.NewToolComponentService()
	component, err := componentService.CreateComponent(ctx, userID, req.Name, req.Description, req.Type, req.AssetID, req.ServiceURL, req.ParamDesc, req.CronExpression)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to create component: %v", err)
		c.JSON(hzconsts.StatusOK, ToolComponentResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Component created: componentID=%s, userID=%s", component.ComponentID, userID)
	c.JSON(hzconsts.StatusOK, ToolComponentResponse{
		Status: "ok",
		Data:   component,
	})
}

// ListToolComponents 列出用户的所有工具组件
// GET /api/tool-component/list
func ListToolComponents(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, ToolComponentResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	componentService := service.NewToolComponentService()
	components, err := componentService.ListComponents(ctx, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to list components: %v", err)
		c.JSON(hzconsts.StatusOK, ToolComponentResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Listed components: userID=%s, count=%d", userID, len(components))
	c.JSON(hzconsts.StatusOK, ToolComponentResponse{
		Status: "ok",
		Data:   components,
	})
}

// GetToolComponent 获取工具组件详情
// GET /api/tool-component/:componentId
func GetToolComponent(ctx context.Context, c *app.RequestContext) {
	componentID := c.Param("componentId")
	if componentID == "" {
		c.JSON(hzconsts.StatusBadRequest, ToolComponentResponse{
			Status: "error",
			Msg:    "ComponentID is required",
		})
		return
	}

	componentService := service.NewToolComponentService()
	component, err := componentService.GetComponent(ctx, componentID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get component: %v", err)
		c.JSON(hzconsts.StatusOK, ToolComponentResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(hzconsts.StatusOK, ToolComponentResponse{
		Status: "ok",
		Data:   component,
	})
}

// UpdateToolComponent 更新工具组件信息
// PUT /api/tool-component/:componentId
func UpdateToolComponent(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, ToolComponentResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	componentID := c.Param("componentId")
	if componentID == "" {
		c.JSON(hzconsts.StatusBadRequest, ToolComponentResponse{
			Status: "error",
			Msg:    "ComponentID is required",
		})
		return
	}

	var req UpdateToolComponentRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid request parameters: %v", err)
		c.JSON(hzconsts.StatusBadRequest, ToolComponentResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	componentService := service.NewToolComponentService()
	component, err := componentService.UpdateComponent(ctx, componentID, userID, req.Name, req.Description, req.AssetID, req.ServiceURL, req.ParamDesc, req.CronExpression)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to update component: %v", err)
		c.JSON(hzconsts.StatusOK, ToolComponentResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Component updated: componentID=%s, userID=%s", componentID, userID)
	c.JSON(hzconsts.StatusOK, ToolComponentResponse{
		Status: "ok",
		Data:   component,
	})
}

// DeleteToolComponent 删除工具组件
// DELETE /api/tool-component/:componentId
func DeleteToolComponent(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, ToolComponentResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	componentID := c.Param("componentId")
	if componentID == "" {
		c.JSON(hzconsts.StatusBadRequest, ToolComponentResponse{
			Status: "error",
			Msg:    "ComponentID is required",
		})
		return
	}

	componentService := service.NewToolComponentService()
	err := componentService.DeleteComponent(ctx, componentID, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to delete component: %v", err)
		c.JSON(hzconsts.StatusOK, ToolComponentResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Component deleted: componentID=%s, userID=%s", componentID, userID)
	c.JSON(hzconsts.StatusOK, ToolComponentResponse{
		Status: "ok",
		Msg:    "Component deleted successfully",
	})
}
