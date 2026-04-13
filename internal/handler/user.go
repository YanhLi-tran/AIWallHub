package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", 10))

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
