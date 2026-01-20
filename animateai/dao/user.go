package dao

import (
	"github.com/AnimateAIPlatform/animate-ai/common/db"
	"github.com/AnimateAIPlatform/animate-ai/models"

	"gorm.io/gorm"
)

// UserDAO 用户 DAO
type UserDAO struct {
	db *gorm.DB
}

// NewUserDAO 使用默认数据库连接创建用户 DAO
func NewUserDAO() *UserDAO {
	return &UserDAO{db: db.DB}
}

// NewUserDAOWithDB 使用指定的数据库连接创建用户 DAO
func NewUserDAOWithDB(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

// Create 插入新用户
func (dao *UserDAO) Create(user *models.User) error {
	return dao.db.Create(user).Error
}

// Update 更新用户
func (dao *UserDAO) Update(user *models.User) error {
	return dao.db.Save(user).Error
}

// Delete 软删除用户
func (dao *UserDAO) Delete(user *models.User) error {
	return dao.db.Delete(user).Error
}

// GetByID 根据ID查询用户
func (dao *UserDAO) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := dao.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUserName 根据用户名查询用户
func (dao *UserDAO) GetByUserName(userName string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("user_name = ? AND deleted_at IS NULL", userName).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByAccountID 根据账户ID查询用户
func (dao *UserDAO) GetByAccountID(accountID string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("account_id = ? AND deleted_at IS NULL", accountID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByName 根据昵称查询用户
func (dao *UserDAO) GetByName(name string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("name = ? AND deleted_at IS NULL", name).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListAll 查询所有用户（未删除的）
func (dao *UserDAO) ListAll() ([]models.User, error) {
	var users []models.User
	err := dao.db.Where("deleted_at IS NULL").Find(&users).Error
	return users, err
}

// UpdateUserInfo 更新用户信息（不更新密码、状态和账户ID）
func (dao *UserDAO) UpdateUserInfo(user *models.User) error {
	// 只更新用户信息相关字段，不更新密码、状态和账户ID（账户ID在注册时生成后不可修改）
	updates := map[string]interface{}{
		"user_name":         user.UserName,
		"name":              user.Name,
		"email":             user.Email,
		"avatar":            user.Avatar,
		"job":               user.Job,
		"job_name":          user.JobName,
		"organization":      user.Organization,
		"organization_name": user.OrganizationName,
		"location":          user.Location,
		"location_name":     user.LocationName,
		"introduction":      user.Introduction,
		"personal_website":  user.PersonalWebsite,
		"verified":          user.Verified,
		"phone_number":      user.PhoneNumber,
		// 注意：account_id 不在更新列表中，账户ID在注册时生成后不可修改
		"address":           user.Address,
		"range_area":        user.RangeArea,
	}
	return dao.db.Model(user).Updates(updates).Error
}

// HardDelete 硬删除用户（物理删除，不是软删除）
func (dao *UserDAO) HardDelete(user *models.User) error {
	return dao.db.Unscoped().Delete(user).Error
}

