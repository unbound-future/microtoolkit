package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/internal/gateway/service"
	"github.com/AnimateAIPlatform/animate-ai/models"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg,omitempty"`
}

// RegisterResponse 注册响应结构
type RegisterResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg,omitempty"`
}

// Login 用户登录接口
func Login(ctx context.Context, c *app.RequestContext) {
	var req LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid login request: %v", err)
		c.JSON(consts.StatusBadRequest, LoginResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	hlog.CtxInfof(ctx, "Login request received: userName=%s, passwordLength=%d", req.UserName, len(req.Password))

	userService := service.NewUserService()
	user, err := userService.Login(ctx, req.UserName, req.Password)
	if err != nil {
		hlog.CtxErrorf(ctx, "Login handler failed: userName=%s, error=%v", req.UserName, err)
		c.JSON(consts.StatusOK, LoginResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Login handler success: userName=%s, userID=%d", user.UserName, user.ID)
	c.JSON(consts.StatusOK, LoginResponse{
		Status: "ok",
	})
}

// Register 用户注册接口
func Register(ctx context.Context, c *app.RequestContext) {
	var req RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid register request: %v", err)
		c.JSON(consts.StatusBadRequest, RegisterResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	userService := service.NewUserService()
	user, err := userService.Register(ctx, req.UserName, req.Password)
	if err != nil {
		hlog.CtxErrorf(ctx, "Register failed: %v", err)
		c.JSON(consts.StatusOK, RegisterResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "User registered: userName=%s, userID=%d", req.UserName, user.ID)
	c.JSON(consts.StatusOK, RegisterResponse{
		Status: "ok",
	})
}

// UserInfoRequest 用户信息请求结构（用于保存）
type UserInfoRequest struct {
	UserName        string `json:"userName,omitempty"`
	Name            string `json:"name,omitempty"`
	Email           string `json:"email,omitempty"`
	Avatar          string `json:"avatar,omitempty"`
	Job             string `json:"job,omitempty"`
	JobName         string `json:"jobName,omitempty"`
	Organization    string `json:"organization,omitempty"`
	OrganizationName string `json:"organizationName,omitempty"`
	Location        string `json:"location,omitempty"`
	LocationName    string `json:"locationName,omitempty"`
	Introduction    string `json:"introduction,omitempty"`
	PersonalWebsite string `json:"personalWebsite,omitempty"`
	Verified        bool   `json:"verified,omitempty"`
	PhoneNumber     string `json:"phoneNumber,omitempty"`
	AccountID       string `json:"accountId,omitempty"`
	Address         string `json:"address,omitempty"`
	RangeArea       string `json:"rangeArea,omitempty"`
}

// UserInfoResponse 用户信息响应结构
type UserInfoResponse struct {
	UserName        string `json:"userName,omitempty"`
	Name            string `json:"name,omitempty"`
	Avatar          string `json:"avatar,omitempty"`
	Email           string `json:"email,omitempty"`
	Job             string `json:"job,omitempty"`
	JobName         string `json:"jobName,omitempty"`
	Organization    string `json:"organization,omitempty"`
	OrganizationName string `json:"organizationName,omitempty"`
	Location        string `json:"location,omitempty"`
	LocationName    string `json:"locationName,omitempty"`
	Introduction    string `json:"introduction,omitempty"`
	PersonalWebsite string `json:"personalWebsite,omitempty"`
	Verified        bool   `json:"verified,omitempty"`
	PhoneNumber     string `json:"phoneNumber,omitempty"`
	AccountID       string `json:"accountId,omitempty"`
	RangeArea       string `json:"rangeArea,omitempty"`
	Address         string `json:"address,omitempty"`
	RegistrationTime string `json:"registrationTime,omitempty"`
	Permissions     interface{} `json:"permissions,omitempty"`
}

// SaveInfoResponse 保存信息响应结构
type SaveInfoResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg,omitempty"`
}

// GetUserInfo 获取用户信息接口
// GET /api/user/userInfo?accountId=xxx 或 GET /api/user/userInfo?userId=xxx
// 也可以使用 userName 参数（向后兼容）
func GetUserInfo(ctx context.Context, c *app.RequestContext) {
	// 优先使用账户ID（账户ID是全局唯一的，不会因为用户名修改而改变）
	accountID := c.Query("accountId")
	// 也支持 account_id 参数（兼容下划线格式）
	if accountID == "" {
		accountID = c.Query("account_id")
	}
	userIDStr := c.Query("userId")
	userName := c.Query("userName")
	
	hlog.CtxInfof(ctx, "GetUserInfo request: accountId=%s, userId=%s, userName=%s", accountID, userIDStr, userName)
	
	userService := service.NewUserService()
	var user *models.User
	var err error

	if accountID != "" {
		// 优先使用账户ID查找用户
		hlog.CtxInfof(ctx, "Looking up user by accountID: %s", accountID)
		user, err = userService.GetUserInfoByAccountID(ctx, accountID)
		if err != nil {
			hlog.CtxErrorf(ctx, "Failed to get user by accountID: %v", err)
		}
	} else if userIDStr != "" {
		var userID uint
		if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
			hlog.CtxErrorf(ctx, "Invalid userID: %s", userIDStr)
			c.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error": "Invalid userID",
			})
			return
		}
		user, err = userService.GetUserInfo(ctx, userID)
	} else if userName != "" {
		// 向后兼容：支持使用用户名查找
		user, err = userService.GetUserInfoByUserName(ctx, userName)
	} else {
		hlog.CtxErrorf(ctx, "Missing accountId, userId or userName parameter")
		c.JSON(consts.StatusBadRequest, map[string]interface{}{
			"error": "Missing accountId, userId or userName parameter",
		})
		return
	}

	if err != nil {
		hlog.CtxErrorf(ctx, "Get user info failed: %v", err)
		c.JSON(consts.StatusOK, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 转换为前端需要的格式
	response := UserInfoResponse{
		UserName:         user.UserName,
		Name:             user.Name,
		Avatar:           user.Avatar,
		Email:            user.Email,
		Job:              user.Job,
		JobName:          user.JobName,
		Organization:     user.Organization,
		OrganizationName: user.OrganizationName,
		Location:         user.Location,
		LocationName:     user.LocationName,
		Introduction:     user.Introduction,
		PersonalWebsite:  user.PersonalWebsite,
		Verified:         user.Verified,
		PhoneNumber:      user.PhoneNumber,
		AccountID:        user.AccountID,
		RangeArea:        user.RangeArea,
		Address:          user.Address,
		RegistrationTime: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 如果没有设置 Name，使用 UserName
	if response.Name == "" {
		response.Name = user.UserName
	}

	// 确保 AccountID 有值，如果没有则使用默认值
	if response.AccountID == "" {
		// 如果数据库中没有 AccountID，生成一个（兼容旧数据）
		timestamp := time.Now().UnixNano()
		data := strconv.FormatInt(timestamp, 10) + "_" + user.UserName
		hash := sha256.Sum256([]byte(data))
		fullHash := hex.EncodeToString(hash[:])
		response.AccountID = fullHash[:50]
	}

	hlog.CtxInfof(ctx, "Get user info success: userName=%s, userID=%d, accountID=%s", user.UserName, user.ID, response.AccountID)
	c.JSON(consts.StatusOK, response)
}

// SaveUserInfo 保存用户信息接口
// POST /api/user/saveInfo?accountId=xxx 或 POST /api/user/saveInfo?userId=xxx
func SaveUserInfo(ctx context.Context, c *app.RequestContext) {
	accountID := c.Query("accountId")
	userIDStr := c.Query("userId")

	var req UserInfoRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.CtxErrorf(ctx, "Invalid save info request: %v", err)
		c.JSON(consts.StatusBadRequest, SaveInfoResponse{
			Status: "error",
			Msg:    "Invalid request parameters",
		})
		return
	}

	userService := service.NewUserService()
	var userID uint
	var err error

	if accountID != "" {
		// 优先使用账户ID查找用户（账户ID是全局唯一的，不会因为用户名修改而改变）
		user, err := userService.GetUserInfoByAccountID(ctx, accountID)
		if err != nil {
			hlog.CtxErrorf(ctx, "User not found: accountID=%s", accountID)
			c.JSON(consts.StatusOK, SaveInfoResponse{
				Status: "error",
				Msg:    "User not found",
			})
			return
		}
		userID = user.ID
	} else if userIDStr != "" {
		if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
			hlog.CtxErrorf(ctx, "Invalid userID: %s", userIDStr)
			c.JSON(consts.StatusBadRequest, SaveInfoResponse{
				Status: "error",
				Msg:    "Invalid userID",
			})
			return
		}
	} else {
		hlog.CtxErrorf(ctx, "Missing accountId or userId parameter")
		c.JSON(consts.StatusBadRequest, SaveInfoResponse{
			Status: "error",
			Msg:    "Missing accountId or userId parameter",
		})
		return
	}

	// 转换为 User 模型（不包含 AccountID，账户ID在注册时生成后不可修改）
	userInfo := &models.User{
		UserName:         req.UserName,
		Name:             req.Name,
		Email:            req.Email,
		Avatar:           req.Avatar,
		Job:              req.Job,
		JobName:          req.JobName,
		Organization:     req.Organization,
		OrganizationName: req.OrganizationName,
		Location:         req.Location,
		LocationName:     req.LocationName,
		Introduction:     req.Introduction,
		PersonalWebsite:  req.PersonalWebsite,
		Verified:         req.Verified,
		PhoneNumber:      req.PhoneNumber,
		// 注意：AccountID 不传递，账户ID在注册时生成后不可修改
		Address:          req.Address,
		RangeArea:        req.RangeArea,
	}

	err = userService.UpdateUserInfo(ctx, userID, userInfo)
	if err != nil {
		hlog.CtxErrorf(ctx, "Save user info failed: %v", err)
		c.JSON(consts.StatusOK, SaveInfoResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Save user info success: userID=%d", userID)
	c.JSON(consts.StatusOK, SaveInfoResponse{
		Status: "ok",
	})
}

// DeleteUserRequest 删除用户请求结构
type DeleteUserRequest struct {
	UserID uint `json:"userId" binding:"required"`
}

// DeleteUserResponse 删除用户响应结构
type DeleteUserResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg,omitempty"`
}

// ListUsersResponse 用户列表响应结构
type ListUsersResponse struct {
	Status string        `json:"status"`
	Data   []models.User `json:"data,omitempty"`
	Msg    string        `json:"msg,omitempty"`
}

// DeleteUser 删除用户接口（软删除）
// DELETE /api/user/:userId 或 POST /api/user/delete
func DeleteUser(ctx context.Context, c *app.RequestContext) {
	// 从路径参数获取 userID
	userIDStr := c.Param("userId")
	if userIDStr == "" {
		// 如果没有路径参数，尝试从请求体获取
		var req DeleteUserRequest
		if err := c.BindAndValidate(&req); err == nil {
			userIDStr = fmt.Sprintf("%d", req.UserID)
		}
	}

	if userIDStr == "" {
		hlog.CtxErrorf(ctx, "Missing userId parameter")
		c.JSON(consts.StatusBadRequest, DeleteUserResponse{
			Status: "error",
			Msg:    "Missing userId parameter",
		})
		return
	}

	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		hlog.CtxErrorf(ctx, "Invalid userID: %s", userIDStr)
		c.JSON(consts.StatusBadRequest, DeleteUserResponse{
			Status: "error",
			Msg:    "Invalid userID",
		})
		return
	}

	userService := service.NewUserService()
	err := userService.DeleteUser(ctx, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "Delete user failed: %v", err)
		c.JSON(consts.StatusOK, DeleteUserResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "Delete user success: userID=%d", userID)
	c.JSON(consts.StatusOK, DeleteUserResponse{
		Status: "ok",
	})
}

// ListUsers 获取用户列表接口
// GET /api/user/list
func ListUsers(ctx context.Context, c *app.RequestContext) {
	userService := service.NewUserService()
	users, err := userService.ListUsers(ctx)
	if err != nil {
		hlog.CtxErrorf(ctx, "List users failed: %v", err)
		c.JSON(consts.StatusOK, ListUsersResponse{
			Status: "error",
			Msg:    err.Error(),
		})
		return
	}

	hlog.CtxInfof(ctx, "List users success: count=%d", len(users))
	c.JSON(consts.StatusOK, ListUsersResponse{
		Status: "ok",
		Data:   users,
	})
}
