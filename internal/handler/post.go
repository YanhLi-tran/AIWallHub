package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"AIWallHub/pkg/validator"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CreatePost发布动态
func CreatePost(c *gin.Context) {
	//获取当前登录用户
	rawUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "请先登录",
		})
		return
	}
	userID, ok := rawUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户ID类型错误",
		})
		return
	}

	postType := c.PostForm("type")
	content := strings.TrimSpace(c.PostForm("content"))
	file, fileErr := c.FormFile("media")

	// 类型校验
	if postType != "text" && postType != "image" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "类型只能是 text 或 image",
		})
		return
	}

	//文字类型处理
	if postType == "text" {
		if content == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "文字内容不能为空",
			})
			return
		}
		if len([]rune(content)) > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "文字内容不能超过1000字",
			})
			return
		}

		post := model.Post{
			UserID:    userID,
			Type:      "text",
			Content:   content,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := config.DB.Create(&post).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "发布失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "发布成功",
			"post_id": post.ID,
		})
		return
	}

	//图片类型处理
	if postType == "image" {
		if fileErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请上传图片",
			})
			return
		}

		ok, msg := validator.ValidateImage(file, "post_image")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": msg,
			})
			return
		}

		timestamp := time.Now().Unix()
		filename := strconv.Itoa(int(userID)) + "_" + strconv.FormatInt(timestamp, 10) + "_" + file.Filename
		savePath := "./uploads/" + filename
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "保存图片失败",
			})
			return
		}
		mediaURL := "/uploads/" + filename

		post := model.Post{
			UserID:    userID,
			Type:      "image",
			Content:   content,
			MediaURL:  mediaURL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := config.DB.Create(&post).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "发布失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "发布成功",
			"post_id": post.ID,
		})
		return
	}
}

// GetPosts获取动态列表
func GetPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var posts []model.Post
	var total int64

	config.DB.Model(&model.Post{}).Order("created_at DESC").Count(&total)
	config.DB.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&posts)

	var result []gin.H
	for _, post := range posts {
		var user model.User
		config.DB.First(&user, post.UserID)

		result = append(result, gin.H{
			"id": post.ID,
			"user": gin.H{
				"id":   user.ID,
				"name": user.Name,
			},
			"type":       post.Type,
			"content":    post.Content,
			"media_url":  post.MediaURL,
			"likes":      post.Likes,
			"views":      post.Views,
			"created_at": post.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"list":      result,
	})
}

// GetPost 获取单条动态详情
func GetPost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var post model.Post
	if err := config.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "动态不存在"})
		return
	}

	config.DB.Model(&post).Update("views", post.Views+1)

	var user model.User
	config.DB.First(&user, post.UserID)

	c.JSON(http.StatusOK, gin.H{
		"id": post.ID,
		"user": gin.H{
			"id":   user.ID,
			"name": user.Name,
		},
		"type":       post.Type,
		"content":    post.Content,
		"media_url":  post.MediaURL,
		"likes":      post.Likes,
		"views":      post.Views + 1,
		"created_at": post.CreatedAt,
	})
}

// DeletePost删除动态
func DeletePost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的ID",
		})
		return
	}

	rawUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "请先登录",
		})
		return
	}

	userID, ok := rawUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	var post model.Post
	if err := config.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "动态不存在",
		})
		return
	}

	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除他人的动态"})
		return
	}

	config.DB.Delete(&post)

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetUserPosts获取某个用户的所有动态
func GetUserPosts(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var posts []model.Post
	var total int64

	config.DB.Model(&model.Post{}).Where("user_id = ?", userID).Order("created_at DESC").Count(&total)
	config.DB.Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&posts)

	var result []gin.H
	for _, post := range posts {
		result = append(result, gin.H{
			"id":         post.ID,
			"type":       post.Type,
			"content":    post.Content,
			"media_url":  post.MediaURL,
			"likes":      post.Likes,
			"views":      post.Views,
			"created_at": post.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"list":      result,
	})
}
