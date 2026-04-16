package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CreateComment 发表评论
func CreateComment(c *gin.Context) {
	// 获取当前用户
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

	// 获取动态ID
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的动态ID",
		})
		return
	}

	// 解析请求体
	var json struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数格式错误",
		})
		return
	}

	// 校验内容
	content := strings.TrimSpace(json.Content)
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "评论内容不能为空",
		})
		return
	}
	if len([]rune(content)) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "评论内容不能超过500字",
		})
		return
	}

	// 检查动态是否存在
	var post model.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "动态不存在",
		})
		return
	}

	// 创建评论
	comment := model.Comment{
		UserID:  userID,
		PostID:  uint(postID),
		Content: content,
	}
	if err := config.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "评论失败",
		})
		return
	}

	// 更新动态的评论数
	config.DB.Model(&post).Update("comments", post.Comments+1)

	c.JSON(http.StatusOK, gin.H{
		"message":    "评论成功",
		"comment_id": comment.ID,
	})
}

// DeleteComment 删除评论
func DeleteComment(c *gin.Context) {
	// 获取当前用户
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

	// 获取评论ID
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评论ID",
		})
		return
	}

	// 查找评论
	var comment model.Comment
	if err := config.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评论不存在",
		})
		return
	}

	// 检查权限（只能删自己的）
	if comment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "无权删除他人的评论",
		})
		return
	}

	// 删除评论
	if err := config.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除失败",
		})
		return
	}

	// 更新动态的评论数
	var post model.Post
	config.DB.First(&post, comment.PostID)
	config.DB.Model(&post).Update("comments", post.Comments-1)

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// GetComments 获取评论列表
func GetComments(c *gin.Context) {
	// 获取动态ID
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的动态ID",
		})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var comments []model.Comment
	var total int64

	// 查询评论列表
	config.DB.Model(&model.Comment{}).
		Where("post_id = ?", postID).
		Count(&total)

	config.DB.Where("post_id = ?", postID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&comments)

	// 补充用户信息
	var result []gin.H
	for _, comment := range comments {
		var user model.User
		config.DB.First(&user, comment.UserID)

		result = append(result, gin.H{
			"id": comment.ID,
			"user": gin.H{
				"id":   user.ID,
				"name": user.Name,
			},
			"content":    comment.Content,
			"created_at": comment.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"list":      result,
	})
}
