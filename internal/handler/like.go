package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LikePost 点赞动态
func LikePost(c *gin.Context) {
	rawUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	userID, ok := rawUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的动态ID"})
		return
	}

	var post model.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "动态不存在"})
		return
	}

	var existingLike model.Like
	if err := config.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingLike).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已经点过赞了"})
		return
	}

	like := model.Like{
		UserID: userID,
		PostID: uint(postID),
	}
	result := config.DB.Create(&like)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
		return
	}

	config.DB.Model(&post).Update("likes", post.Likes+1)

	c.JSON(http.StatusOK, gin.H{"message": "点赞成功"})
}

// UnlikePost 取消点赞
func UnlikePost(c *gin.Context) {
	rawUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	userID, ok := rawUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的动态ID"})
		return
	}

	var like model.Like
	if err := config.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "还没有点过赞"})
		return
	}

	result := config.DB.Delete(&like)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败"})
		return
	}

	var post model.Post
	config.DB.First(&post, postID)
	config.DB.Model(&post).Update("likes", post.Likes-1)

	c.JSON(http.StatusOK, gin.H{"message": "取消点赞成功"})
}

// GetPostLikes 获取动态的点赞用户列表
func GetPostLikes(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的动态ID"})
		return
	}

	var post model.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "动态不存在"})
		return
	}

	rawUserID, exists := c.Get("current_user_id")
	var currentUserID uint
	if exists {
		if uid, ok := rawUserID.(uint); ok {
			currentUserID = uid
		}
	}

	if currentUserID != post.UserID && !post.ShowLikes {
		c.JSON(http.StatusForbidden, gin.H{"error": "作者未公开点赞列表"})
		return
	}

	var likes []model.Like
	config.DB.Where("post_id = ?", postID).Find(&likes)

	var result []gin.H
	for _, like := range likes {
		var user model.User
		config.DB.First(&user, like.UserID)
		result = append(result, gin.H{
			"user_id":  user.ID,
			"username": user.Name,
			"liked_at": like.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(result),
		"list":  result,
	})
}

// GetUserLikes 获取用户的点赞列表
func GetUserLikes(c *gin.Context) {
	rawUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	currentUserID, ok := rawUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var user model.User
	if err := config.DB.First(&user, targetUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if currentUserID != uint(targetUserID) && !user.LikesVisible {
		c.JSON(http.StatusForbidden, gin.H{"error": "该用户未公开点赞列表"})
		return
	}

	var likes []model.Like
	config.DB.Where("user_id = ?", targetUserID).Order("created_at DESC").Find(&likes)

	var result []gin.H
	for _, like := range likes {
		var post model.Post
		config.DB.First(&post, like.PostID)

		var mediaURLs []string
		if post.MediaURLs != "" {
			json.Unmarshal([]byte(post.MediaURLs), &mediaURLs)
		}

		result = append(result, gin.H{
			"post_id":         post.ID,
			"type":            post.Type,
			"content":         post.Content,
			"media_urls":      mediaURLs,
			"video_url":       post.VideoURL,
			"video_duration":  post.VideoDuration,
			"video_thumbnail": post.VideoThumbnail,
			"likes":           post.Likes,
			"comments_count":  post.Comments,
			"liked_at":        like.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(result),
		"list":  result,
	})
}
