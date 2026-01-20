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

// AgentFlowResponse 工作流响应
type AgentFlowResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Msg    string      `json:"msg,omitempty"`
}

// CreateAgentFlowRequest 创建工作流请求
type CreateAgentFlowRequest struct {
	Name      string      `json:"name" binding:"required"`        // 工作流名称
	AssetID   string      `json:"asset_id,omitempty"`             // 关联的资产ID（可选）
	TemplateID string     `json:"template_id,omitempty"`          // 工作流模版ID（可选）
	FlowData  interface{} `json:"flow_data" binding:"required"`   // 工作流数据（JSON格式）
}

// UpdateAgentFlowRequest 更新工作流请求
type UpdateAgentFlowRequest struct {
	Name      string      `json:"name" binding:"required"`        // 工作流名称
	AssetID   string      `json:"asset_id,omitempty"`             // 关联的资产ID（可选）
	TemplateID string     `json:"template_id,omitempty"`          // 工作流模版ID（可选）
	FlowData  interface{} `json:"flow_data" binding:"required"`   // 工作流数据（JSON格式）
}

// CreateAgentFlow 创建工作流接口
// POST /api/agent-flow
func CreateAgentFlow(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AgentFlowResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	var req CreateAgentFlowRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid request parameters: %v", err)
		c.JSON(hzconsts.StatusBadRequest, AgentFlowResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	agentFlowService := service.NewAgentFlowService()
	flow, err := agentFlowService.CreateAgentFlow(ctx, userID, req.Name, req.AssetID, req.TemplateID, req.FlowData)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to create agent flow: %v", err)
		c.JSON(hzconsts.StatusOK, AgentFlowResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(hzconsts.StatusOK, AgentFlowResponse{
		Status: "ok",
		Data:   flow,
	})
}

// GetAgentFlow 获取工作流详情接口
// GET /api/agent-flow/:flowId
func GetAgentFlow(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AgentFlowResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	flowID := c.Param("flowId")
	if flowID == "" {
		c.JSON(hzconsts.StatusBadRequest, AgentFlowResponse{
			Status: "error",
			Msg:    "FlowID is required",
		})
		return
	}

	agentFlowService := service.NewAgentFlowService()
	flow, err := agentFlowService.GetAgentFlow(ctx, flowID, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get agent flow: %v", err)
		c.JSON(hzconsts.StatusOK, AgentFlowResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	// 解析 FlowData JSON 字符串为对象
	var flowData interface{}
	if flow.FlowData != "" {
		if err := json.Unmarshal([]byte(flow.FlowData), &flowData); err != nil {
			hlog.CtxWarnf(ctx, "Failed to unmarshal flow data: %v", err)
			flowData = flow.FlowData // 如果解析失败，返回原始字符串
		}
	}

	// 构造返回数据，将 FlowData 转换为对象
	responseData := map[string]interface{}{
		"id":          flow.ID,
		"flow_id":     flow.FlowID,
		"user_id":     flow.UserID,
		"name":        flow.Name,
		"asset_id":    flow.AssetID,
		"template_id": flow.TemplateID,
		"flow_data":   flowData,
		"created_at":  flow.CreatedAt,
		"updated_at":  flow.UpdatedAt,
	}

	c.JSON(hzconsts.StatusOK, AgentFlowResponse{
		Status: "ok",
		Data:   responseData,
	})
}

// UpdateAgentFlow 更新工作流接口
// PUT /api/agent-flow/:flowId
func UpdateAgentFlow(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AgentFlowResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	flowID := c.Param("flowId")
	if flowID == "" {
		c.JSON(hzconsts.StatusBadRequest, AgentFlowResponse{
			Status: "error",
			Msg:    "FlowID is required",
		})
		return
	}

	var req UpdateAgentFlowRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid request parameters: %v", err)
		c.JSON(hzconsts.StatusBadRequest, AgentFlowResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	agentFlowService := service.NewAgentFlowService()
	flow, err := agentFlowService.UpdateAgentFlow(ctx, flowID, userID, req.Name, req.AssetID, req.TemplateID, req.FlowData)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to update agent flow: %v", err)
		c.JSON(hzconsts.StatusOK, AgentFlowResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(hzconsts.StatusOK, AgentFlowResponse{
		Status: "ok",
		Data:   flow,
	})
}

// DeleteAgentFlow 删除工作流接口
// DELETE /api/agent-flow/:flowId
func DeleteAgentFlow(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AgentFlowResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	flowID := c.Param("flowId")
	if flowID == "" {
		c.JSON(hzconsts.StatusBadRequest, AgentFlowResponse{
			Status: "error",
			Msg:    "FlowID is required",
		})
		return
	}

	agentFlowService := service.NewAgentFlowService()
	err := agentFlowService.DeleteAgentFlow(ctx, flowID, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to delete agent flow: %v", err)
		c.JSON(hzconsts.StatusOK, AgentFlowResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(hzconsts.StatusOK, AgentFlowResponse{
		Status: "ok",
		Msg:    "Agent flow deleted successfully",
	})
}

// ListAgentFlows 列出用户的所有工作流接口
// GET /api/agent-flow/list
func ListAgentFlows(ctx context.Context, c *app.RequestContext) {
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AgentFlowResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	agentFlowService := service.NewAgentFlowService()
	flows, err := agentFlowService.ListAgentFlows(ctx, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to list agent flows: %v", err)
		c.JSON(hzconsts.StatusOK, AgentFlowResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	// 转换数据，解析 FlowData JSON 字符串为对象
	responseData := make([]map[string]interface{}, 0, len(flows))
	for _, flow := range flows {
		var flowData interface{}
		if flow.FlowData != "" {
			if err := json.Unmarshal([]byte(flow.FlowData), &flowData); err != nil {
				hlog.CtxWarnf(ctx, "Failed to unmarshal flow data for flowID=%s: %v", flow.FlowID, err)
				flowData = flow.FlowData // 如果解析失败，返回原始字符串
			}
		}

		responseData = append(responseData, map[string]interface{}{
			"id":          flow.ID,
			"flow_id":     flow.FlowID,
			"user_id":     flow.UserID,
			"name":        flow.Name,
			"asset_id":    flow.AssetID,
			"template_id": flow.TemplateID,
			"flow_data":   flowData,
			"created_at":  flow.CreatedAt,
			"updated_at":  flow.UpdatedAt,
		})
	}

	c.JSON(hzconsts.StatusOK, AgentFlowResponse{
		Status: "ok",
		Data:   responseData,
	})
}



