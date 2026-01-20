package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/internal/gateway/service"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hzconsts "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// WorkflowTemplateResponse 工作流模版响应
type WorkflowTemplateResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Msg    string      `json:"msg,omitempty"`
}

// CreateWorkflowTemplateRequest 创建工作流模版请求
type CreateWorkflowTemplateRequest struct {
	Name         string      `json:"name" binding:"required"`        // 模版名称
	Description  string      `json:"description,omitempty"`         // 模版描述（可选）
	AssetID      string      `json:"asset_id,omitempty"`             // 关联的资产ID（可选）
	TemplateData interface{} `json:"template_data" binding:"required"` // 模版数据（JSON格式）
}

// UpdateWorkflowTemplateRequest 更新工作流模版请求
type UpdateWorkflowTemplateRequest struct {
	Name         string      `json:"name" binding:"required"`        // 模版名称
	Description  string      `json:"description,omitempty"`          // 模版描述（可选）
	AssetID      string      `json:"asset_id,omitempty"`             // 关联的资产ID（可选）
	TemplateData interface{} `json:"template_data" binding:"required"` // 模版数据（JSON格式）
}

// CreateWorkflowTemplate 创建工作流模版接口
// POST /api/workflow-template
func CreateWorkflowTemplate(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%v", userIDValue)

	var req CreateWorkflowTemplateRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid request parameters: %v", err)
		c.JSON(hzconsts.StatusBadRequest, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	templateService := service.NewWorkflowTemplateService()
	template, err := templateService.CreateWorkflowTemplate(ctx, userID, req.Name, req.Description, req.AssetID, req.TemplateData)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to create workflow template: %v", err)
		c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
		Status: "ok",
		Data:   template,
	})
}

// GetWorkflowTemplate 获取工作流模版详情接口
// GET /api/workflow-template/:templateId
func GetWorkflowTemplate(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%v", userIDValue)

	templateID := c.Param("templateId")
	if templateID == "" {
		c.JSON(hzconsts.StatusBadRequest, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "TemplateID is required",
		})
		return
	}

	templateService := service.NewWorkflowTemplateService()
	template, err := templateService.GetWorkflowTemplate(ctx, templateID, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get workflow template: %v", err)
		c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	// 解析 TemplateData JSON 字符串为对象
	var templateData interface{}
	if template.TemplateData != "" {
		if err := json.Unmarshal([]byte(template.TemplateData), &templateData); err != nil {
			hlog.CtxWarnf(ctx, "Failed to unmarshal template data: %v", err)
			templateData = template.TemplateData // 如果解析失败，返回原始字符串
		}
	}

	// 构造返回数据，将 TemplateData 转换为对象
	responseData := map[string]interface{}{
		"id":            template.ID,
		"template_id":   template.TemplateID,
		"user_id":       template.UserID,
		"name":          template.Name,
		"description":  template.Description,
		"asset_id":      template.AssetID,
		"template_data": templateData,
		"created_at":    template.CreatedAt,
		"updated_at":    template.UpdatedAt,
	}

	c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
		Status: "ok",
		Data:   responseData,
	})
}

// UpdateWorkflowTemplate 更新工作流模版接口
// PUT /api/workflow-template/:templateId
func UpdateWorkflowTemplate(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%v", userIDValue)

	templateID := c.Param("templateId")
	if templateID == "" {
		c.JSON(hzconsts.StatusBadRequest, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "TemplateID is required",
		})
		return
	}

	var req UpdateWorkflowTemplateRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid request parameters: %v", err)
		c.JSON(hzconsts.StatusBadRequest, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	templateService := service.NewWorkflowTemplateService()
	template, err := templateService.UpdateWorkflowTemplate(ctx, templateID, userID, req.Name, req.Description, req.AssetID, req.TemplateData)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to update workflow template: %v", err)
		c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
		Status: "ok",
		Data:   template,
	})
}

// DeleteWorkflowTemplate 删除工作流模版接口
// DELETE /api/workflow-template/:templateId
func DeleteWorkflowTemplate(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%v", userIDValue)

	templateID := c.Param("templateId")
	if templateID == "" {
		c.JSON(hzconsts.StatusBadRequest, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "TemplateID is required",
		})
		return
	}

	templateService := service.NewWorkflowTemplateService()
	err := templateService.DeleteWorkflowTemplate(ctx, templateID, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to delete workflow template: %v", err)
		c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
		Status: "ok",
		Msg:    "Workflow template deleted successfully",
	})
}

// ListWorkflowTemplates 列出用户的所有工作流模版接口
// GET /api/workflow-template/list
func ListWorkflowTemplates(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, WorkflowTemplateResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%v", userIDValue)

	templateService := service.NewWorkflowTemplateService()
	templates, err := templateService.ListWorkflowTemplates(ctx, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to list workflow templates: %v", err)
		c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	// 转换数据，解析 TemplateData JSON 字符串为对象（可选，列表可能不需要完整数据）
	responseData := make([]map[string]interface{}, 0, len(templates))
	for _, template := range templates {
		responseData = append(responseData, map[string]interface{}{
			"id":           template.ID,
			"template_id":  template.TemplateID,
			"user_id":     template.UserID,
			"name":        template.Name,
			"description": template.Description,
			"asset_id":    template.AssetID,
			"created_at":  template.CreatedAt,
			"updated_at":  template.UpdatedAt,
		})
	}

	c.JSON(hzconsts.StatusOK, WorkflowTemplateResponse{
		Status: "ok",
		Data:   responseData,
	})
}
