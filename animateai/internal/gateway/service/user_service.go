package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/dao"
	"github.com/AnimateAIPlatform/animate-ai/models"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	userDAO *dao.UserDAO
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		userDAO: dao.NewUserDAO(),
	}
}

// NewUserServiceWithDB 使用指定的数据库连接创建用户服务
func NewUserServiceWithDB(db *gorm.DB) *UserService {
	return &UserService{
		userDAO: dao.NewUserDAOWithDB(db),
	}
}

// Register 用户注册，返回创建的用户
func (s *UserService) Register(ctx context.Context, userName, password string) (*models.User, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userDAO.GetByUserName(userName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		hlog.CtxErrorf(ctx, "Failed to check user existence: %v", err)
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to hash password: %v", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 生成全局唯一的账户ID：使用SHA256 hash算法
	// 将当前时间戳（纳秒）+ 用户名进行hash，确保唯一性
	timestamp := time.Now().UnixNano()
	data := strconv.FormatInt(timestamp, 10) + "_" + userName
	hash := sha256.Sum256([]byte(data))
	fullHash := hex.EncodeToString(hash[:])
	// 截取前50个字符（数据库字段长度为varchar(50)）
	// 50个十六进制字符仍然具有极高的唯一性
	accountID := fullHash[:50]

	// 创建用户
	user := &models.User{
		UserName:  userName,
		Password:  string(hashedPassword),
		Status:    1,         // 正常状态
		AccountID: accountID, // 全局唯一的账户ID
	}

	err = s.userDAO.Create(user)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 等待一小段时间确保数据库事务已提交
	time.Sleep(50 * time.Millisecond)

	// 重新查询用户以确保获取完整信息（包括 ID）
	createdUser, err := s.userDAO.GetByUserName(userName)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to get created user: %v", err)
		return nil, fmt.Errorf("failed to get created user: %w", err)
	}

	hlog.CtxInfof(ctx, "User registered successfully: userName=%s, userID=%d", userName, createdUser.ID)
	return createdUser, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, userName, password string) (*models.User, error) {
	// 记录登录请求信息（密码只记录长度，不记录实际内容）
	passwordLen := len(password)
	hlog.CtxInfof(ctx, "Login attempt: userName=%s, passwordLength=%d", userName, passwordLen)

	// 查询用户
	user, err := s.userDAO.GetByUserName(userName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hlog.CtxWarnf(ctx, "Login failed: user not found, userName=%s", userName)
			return nil, fmt.Errorf("username or password incorrect")
		}
		hlog.CtxErrorf(ctx, "Login failed: database error, userName=%s, error=%v", userName, err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	hlog.CtxInfof(ctx, "User found: userName=%s, userID=%d, status=%d, hasPassword=%v",
		user.UserName, user.ID, user.Status, user.Password != "")

	// 检查用户状态
	if user.Status != 1 {
		hlog.CtxWarnf(ctx, "Login failed: user account disabled, userName=%s, status=%d", userName, user.Status)
		return nil, fmt.Errorf("user account is disabled")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		hlog.CtxWarnf(ctx, "Login failed: password mismatch, userName=%s, error=%v", userName, err)
		return nil, fmt.Errorf("username or password incorrect")
	}

	hlog.CtxInfof(ctx, "User logged in successfully: userName=%s, userID=%d", userName, user.ID)
	return user, nil
}

// GetUserInfo 获取用户信息（根据用户ID）
func (s *UserService) GetUserInfo(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		hlog.CtxErrorf(ctx, "Failed to get user info: %v", err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	return user, nil
}

// GetUserInfoByUserName 根据用户名获取用户信息
func (s *UserService) GetUserInfoByUserName(ctx context.Context, userName string) (*models.User, error) {
	user, err := s.userDAO.GetByUserName(userName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		hlog.CtxErrorf(ctx, "Failed to get user info: %v", err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	return user, nil
}

// GetUserInfoByAccountID 根据账户ID获取用户信息
func (s *UserService) GetUserInfoByAccountID(ctx context.Context, accountID string) (*models.User, error) {
	user, err := s.userDAO.GetByAccountID(accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		hlog.CtxErrorf(ctx, "Failed to get user info by accountID: %v", err)
		return nil, fmt.Errorf("failed to get user info by accountID: %w", err)
	}
	return user, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(ctx context.Context, userID uint, userInfo *models.User) error {
	// 先查询用户是否存在
	existingUser, err := s.userDAO.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		hlog.CtxErrorf(ctx, "Failed to get user: %v", err)
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 如果提供了新用户名，检查是否已被使用
	if userInfo.UserName != "" && userInfo.UserName != existingUser.UserName {
		// 检查新用户名是否已被其他用户使用
		otherUser, err := s.userDAO.GetByUserName(userInfo.UserName)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			hlog.CtxErrorf(ctx, "Failed to check username availability: %v", err)
			return fmt.Errorf("failed to check username availability: %w", err)
		}
		if otherUser != nil && otherUser.ID != userID {
			return fmt.Errorf("username already exists")
		}
		existingUser.UserName = userInfo.UserName
	}

	// 如果提供了新昵称，检查是否已被使用
	if userInfo.Name != "" && userInfo.Name != existingUser.Name {
		// 检查新昵称是否已被其他用户使用
		otherUser, err := s.userDAO.GetByName(userInfo.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			hlog.CtxErrorf(ctx, "Failed to check name availability: %v", err)
			return fmt.Errorf("failed to check name availability: %w", err)
		}
		if otherUser != nil && otherUser.ID != userID {
			return fmt.Errorf("name already exists")
		}
		existingUser.Name = userInfo.Name
	}

	// 更新用户信息字段（不更新账户ID，账户ID在注册时生成后不可修改）
	existingUser.Email = userInfo.Email
	existingUser.Avatar = userInfo.Avatar
	existingUser.Job = userInfo.Job
	existingUser.JobName = userInfo.JobName
	existingUser.Organization = userInfo.Organization
	existingUser.OrganizationName = userInfo.OrganizationName
	existingUser.Location = userInfo.Location
	existingUser.LocationName = userInfo.LocationName
	existingUser.Introduction = userInfo.Introduction
	existingUser.PersonalWebsite = userInfo.PersonalWebsite
	existingUser.Verified = userInfo.Verified
	existingUser.PhoneNumber = userInfo.PhoneNumber
	// 注意：AccountID 不更新，账户ID在注册时生成后不可修改
	existingUser.Address = userInfo.Address
	existingUser.RangeArea = userInfo.RangeArea

	err = s.userDAO.UpdateUserInfo(existingUser)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to update user info: %v", err)
		return fmt.Errorf("failed to update user info: %w", err)
	}

	hlog.CtxInfof(ctx, "User info updated successfully: userID=%d", userID)
	return nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(ctx context.Context, userID uint) error {
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		hlog.CtxErrorf(ctx, "Failed to get user: %v", err)
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 软删除
	err = s.userDAO.Delete(user)
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to delete user: %v", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	hlog.CtxInfof(ctx, "User deleted successfully (soft delete): userID=%d", userID)
	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.userDAO.ListAll()
	if err != nil {
		hlog.CtxErrorf(ctx, "Failed to list users: %v", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	hlog.CtxInfof(ctx, "List users successfully: count=%d", len(users))
	return users, nil
}
