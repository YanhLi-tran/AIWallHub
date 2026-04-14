package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"AIWallHub/pkg/cache"
	"AIWallHub/pkg/crypto"
	"AIWallHub/pkg/email"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SendVerifyCode 发送邮箱验证码
func SendVerifyCode(c *gin.Context) {
	var json struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数格式错误",
		})
		return
	}

	// 校验邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(json.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "邮箱格式不正确",
		})
		return
	}

	// 生成验证码
	code := email.GenerateCode()

	// 存储验证码
	cache.Set(json.Email, code)

	// 发送邮件
	if err := email.SendVerificationCode(json.Email, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "发送验证码失败" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "验证码已发送",
	})
}

// 获取单个用户信息
func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	var user model.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	// 获取当前登录用户
	currentUserID, _ := c.Get("current_user_id")

	// 判断是否返回邮箱
	var email string
	if currentUserID == user.ID || user.EmailVisible {
		email = user.Email
	} else {
		email = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"email":      email,
		"avatar":     user.Avatar,
		"bio":        user.Bio,
		"created_at": user.CreatedAt,
	})
}

// 获取用户列表（分页）
func GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var users []model.User
	var total int64

	config.DB.Model(&model.User{}).Count(&total)
	config.DB.Offset(offset).Limit(pageSize).Find(&users)

	var result []gin.H
	for _, user := range users {
		// 根据用户设置决定是否显示邮箱
		email := ""
		if user.EmailVisible {
			email = user.Email
		} else {
			email = ""
		}
		result = append(result, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      email,
			"avatar":     user.Avatar,
			"bio":        user.Bio,
			"created_at": user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"list":      result,
	})
}

// 更新用户信息
func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	//获取当前登录用户
	currentUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "请先登录",
		})
		return
	}

	if currentUserID != uint(id) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "无权修改他人信息",
		})
		return
	}

	var json struct {
		Name         string `json:"name"`
		Avatar       string `json:"avatar"`
		Bio          string `json:"bio"`
		EmailVisible bool   `json:"email_visible"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数格式错误",
		})
		return
	}

	//构建更新数据
	updates := map[string]interface{}{
		"name":          json.Name,
		"avatar":        json.Avatar,
		"bio":           json.Bio,
		"email_visible": json.EmailVisible,
	}

	result := config.DB.Model(&model.User{}).Where("id=?", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新失败" + " " + result.Error.Error(),
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}

// UpdatePassword 修改密码
func UpdatePassword(c *gin.Context) {
	// 获取当前登录用户
	currentUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	var json struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	// 验证旧密码
	var user model.User
	if err := config.DB.First(&user, currentUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if !crypto.CheckPassword(json.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "原密码错误"})
		return
	}

	// 新密码校验
	if len(json.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码至少需要6个字符"})
		return
	}

	// 加密新密码
	hashedPassword, err := crypto.HashPassword(json.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 更新密码
	result := config.DB.Model(&model.User{}).Where("id = ?", currentUserID).Update("password", hashedPassword)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修改密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// UpdateEmail 修改邮箱
func UpdateEmail(c *gin.Context) {
	// 获取当前登录用户
	currentUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	var json struct {
		Password string `json:"password"`
		NewEmail string `json:"new_email"`
		Code     string `json:"code"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	// 验证验证码
	savedCode, ok := cache.Get(json.NewEmail)
	if !ok || savedCode != json.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
		return
	}

	// 验证密码
	var user model.User
	if err := config.DB.First(&user, currentUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if !crypto.CheckPassword(json.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}

	// 检查新邮箱是否已被注册
	var existingUser model.User
	if err := config.DB.Where("email = ?", json.NewEmail).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "该邮箱已被注册"})
		return
	}

	// 更新邮箱
	result := config.DB.Model(&model.User{}).Where("id = ?", currentUserID).Update("email", json.NewEmail)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修改邮箱失败"})
		return
	}

	//删除验证码
	cache.Delete(json.NewEmail)

	c.JSON(http.StatusOK, gin.H{"message": "邮箱修改成功"})
}

// 删除用户
func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	//获取当前登录用户
	currentUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "请先登录",
		})
		return
	}

	if currentUserID != uint(id) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "无权删除他人账户",
		})
		return
	}

	//验证密码防止误删
	var json struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&json); err == nil && json.Password != "" {
		var user model.User
		if err := config.DB.First(&user, id).Error; err == nil {
			if !crypto.CheckPassword(json.Password, user.Password) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "密码错误，无法删除",
				})
				return
			}
		}
	}

	result := config.DB.Delete(&model.User{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除失败" + " " + result.Error.Error(),
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "成功删除",
	})
}
