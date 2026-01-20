package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/internal/gateway/service"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hzconsts "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// UploadAssetRequest 上传资产请求（文件上传）
type UploadAssetRequest struct {
	Name        string `form:"name" binding:"required"`         // 资产名称
	Description string `form:"description"`                     // 资产描述（可选）
}

// AddAssetByURLRequest 通过URL添加资产请求
type AddAssetByURLRequest struct {
	Name        string `json:"name" binding:"required"`         // 资产名称
	Description string `json:"description"`                     // 资产描述（可选）
	URL         string `json:"url" binding:"required"`          // 资产URL
}

// UpdateAssetRequest 更新资产请求
type UpdateAssetRequest struct {
	Name        string `json:"name" binding:"required"`         // 资产名称
	Description string `json:"description"`                     // 资产描述（可选）
	URL         string `json:"url"`                             // 资产URL（仅URL类型资产可修改）
}

// AssetResponse 资产响应
type AssetResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Msg    string      `json:"msg,omitempty"`
}

// UploadAsset 上传资产文件接口
// POST /api/asset/upload
func UploadAsset(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AssetResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get uploaded file: %v", err)
		c.JSON(hzconsts.StatusBadRequest, AssetResponse{
			Status: "error",
			Msg:    "File is required",
		})
		return
	}

	// 获取表单数据（multipart/form-data）
	name := c.PostForm("name")
	description := c.PostForm("description")
	
	if name == "" {
		// 如果没有提供name，使用文件名
		name = fileHeader.Filename
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to open uploaded file: %v", err)
		c.JSON(hzconsts.StatusInternalServerError, AssetResponse{
			Status: "error",
			Msg:    "Failed to process uploaded file",
		})
		return
	}
	defer file.Close()

	// 获取文件信息
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	fileSize := fileHeader.Size

	// 上传文件到COS并保存到数据库
	assetService := service.NewAssetService()
	asset, err := assetService.UploadAsset(
		ctx,
		userID,
		name,
		description,
		file,
		fileHeader.Filename,
		contentType,
		fileSize,
	)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to upload asset: userID=%s, fileName=%s, error=%v", userID, fileHeader.Filename, err)
		// 返回更友好的错误信息
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "no storage config found") {
			errorMsg = "Object storage is not configured. Please configure COS storage first."
		} else if strings.Contains(errorMsg, "failed to create OSS client") {
			errorMsg = "Failed to initialize object storage client. Please check storage configuration."
		} else if strings.Contains(errorMsg, "failed to upload file to COS") {
			errorMsg = "Failed to upload file to object storage. Please check storage configuration and network."
		}
		c.JSON(hzconsts.StatusOK, AssetResponse{
			Status: "error",
			Msg:    errorMsg,
		})
		return
	}

	hlog.CtxInfof(ctx, "Asset uploaded successfully: assetID=%s, userID=%s", asset.AssetID, userID)
	c.JSON(hzconsts.StatusOK, AssetResponse{
		Status: "ok",
		Data:   asset,
	})
}

// AddAssetByURL 通过URL添加资产接口
// POST /api/asset/add-by-url
func AddAssetByURL(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AssetResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	var req AddAssetByURLRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid request parameters: %v", err)
		c.JSON(hzconsts.StatusBadRequest, AssetResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	// 添加资产到数据库
	assetService := service.NewAssetService()
	asset, err := assetService.AddAssetByURL(ctx, userID, req.Name, req.Description, req.URL)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to add asset by URL: %v", err)
		c.JSON(hzconsts.StatusOK, AssetResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Asset added by URL: assetID=%s, userID=%s", asset.AssetID, userID)
	c.JSON(hzconsts.StatusOK, AssetResponse{
		Status: "ok",
		Data:   asset,
	})
}

// ListAssets 列出用户的所有资产
// GET /api/asset/list
func ListAssets(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AssetResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	assetService := service.NewAssetService()
	assets, err := assetService.ListAssets(ctx, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to list assets: %v", err)
		c.JSON(hzconsts.StatusOK, AssetResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Listed assets: userID=%s, count=%d", userID, len(assets))
	c.JSON(hzconsts.StatusOK, AssetResponse{
		Status: "ok",
		Data:   assets,
	})
}

// GetAsset 获取资产详情
// GET /api/asset/:assetId
func GetAsset(ctx context.Context, c *app.RequestContext) {
	assetID := c.Param("assetId")
	if assetID == "" {
		c.JSON(hzconsts.StatusBadRequest, AssetResponse{
			Status: "error",
			Msg:    "AssetID is required",
		})
		return
	}

	assetService := service.NewAssetService()
	asset, err := assetService.GetAsset(ctx, assetID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get asset: %v", err)
		c.JSON(hzconsts.StatusOK, AssetResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(hzconsts.StatusOK, AssetResponse{
		Status: "ok",
		Data:   asset,
	})
}

// GeneratePresignedURL 生成资产的预签名下载链接
// GET /api/asset/:assetId/presigned-url
func GeneratePresignedURL(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AssetResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	assetID := c.Param("assetId")
	if assetID == "" {
		c.JSON(hzconsts.StatusBadRequest, AssetResponse{
			Status: "error",
			Msg:    "AssetID is required",
		})
		return
	}

	assetService := service.NewAssetService()
	presignedURL, err := assetService.GeneratePresignedURL(ctx, assetID, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to generate presigned URL: %v", err)
		c.JSON(hzconsts.StatusOK, AssetResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Generated presigned URL: assetID=%s, userID=%s", assetID, userID)
	c.JSON(hzconsts.StatusOK, AssetResponse{
		Status: "ok",
		Data: map[string]interface{}{
			"url": presignedURL,
		},
	})
}

// UpdateAsset 更新资产信息
// PUT /api/asset/:assetId
func UpdateAsset(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AssetResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	assetID := c.Param("assetId")
	if assetID == "" {
		c.JSON(hzconsts.StatusBadRequest, AssetResponse{
			Status: "error",
			Msg:    "AssetID is required",
		})
		return
	}

	var req UpdateAssetRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid request parameters: %v", err)
		c.JSON(hzconsts.StatusBadRequest, AssetResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	assetService := service.NewAssetService()
	asset, err := assetService.UpdateAsset(ctx, assetID, userID, req.Name, req.Description, req.URL)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to update asset: %v", err)
		c.JSON(hzconsts.StatusOK, AssetResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Asset updated: assetID=%s, userID=%s", assetID, userID)
	c.JSON(hzconsts.StatusOK, AssetResponse{
		Status: "ok",
		Data:   asset,
	})
}

// DeleteAsset 删除资产
// DELETE /api/asset/:assetId
func DeleteAsset(ctx context.Context, c *app.RequestContext) {
	// 从 context 中获取用户信息
	userIDValue := ctx.Value(consts.UserIDKey)
	if userIDValue == nil {
		hlog.CtxErrorf(ctx, "UserID not found in context")
		c.JSON(hzconsts.StatusUnauthorized, AssetResponse{
			Status: "error",
			Msg:    "Unauthorized: UserID not found",
		})
		return
	}
	userID := fmt.Sprintf("%d", userIDValue)

	assetID := c.Param("assetId")
	if assetID == "" {
		c.JSON(hzconsts.StatusBadRequest, AssetResponse{
			Status: "error",
			Msg:    "AssetID is required",
		})
		return
	}

	assetService := service.NewAssetService()
	err := assetService.DeleteAsset(ctx, assetID, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to delete asset: %v", err)
		c.JSON(hzconsts.StatusOK, AssetResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Asset deleted: assetID=%s, userID=%s", assetID, userID)
	c.JSON(hzconsts.StatusOK, AssetResponse{
		Status: "ok",
		Msg:    "Asset deleted successfully",
	})
}

