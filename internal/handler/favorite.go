package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FavoritePost 收藏
func FavoritePost(c *gin.Context) {
	rawUserID, exists := c.Get(
		"current_user_id",
	)
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

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的动态ID",
		})
		return
	}

	var post model.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "动态不存在",
		})
		return
	}

	var existing model.Favorite
	if err := config.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "已经收藏过了",
		})
		return
	}

	favorite := model.Favorite{
		UserID: userID,
		PostID: uint(postID),
	}
	if err := config.DB.Create(&favorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "收藏失败",
		})
		return
	}

	config.DB.Model(&post).Update("favorites", post.Favorites+1)

	c.JSON(http.StatusOK, gin.H{
		"message": "收藏成功",
	})
}

// UnfavoritePost 取消收藏
func UnfavoritePost(c *gin.Context) {
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

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的动态ID",
		})
		return
	}

	var favorite model.Favorite
	if err := config.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&favorite).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "还没有收藏过",
		})
		return
	}

	if err := config.DB.Delete(&favorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "取消收藏失败",
		})
		return
	}

	var post model.Post
	config.DB.First(&post, postID)
	config.DB.Model(&post).Update("favorites", post.Favorites-1)

	c.JSON(http.StatusOK, gin.H{
		"message": "取消收藏成功",
	})
}

// GetFavorites 获取用户的收藏列表（由用户自己决定是否公开）
func GetFavorites(c *gin.Context) {
	// 获取当前用户
	rawUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "请先登录",
		})
		return
	}
	currentUserID, ok := rawUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户ID类型错误",
		})
		return
	}

	// 获取要查看的用户ID
	targetUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	// 获取用户信息
	var user model.User
	if err := config.DB.First(&user, targetUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	// 判断是否公开：查看自己 OR 对方设置了公开
	if currentUserID != uint(targetUserID) && !user.FavoritesVisible {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "该用户未公开收藏列表",
		})
		return
	}

	// 查询收藏记录
	var favorites []model.Favorite
	config.DB.Where("user_id = ?", targetUserID).Order("created_at DESC").Find(&favorites)

	var result []gin.H
	for _, fav := range favorites {
		var post model.Post
		config.DB.First(&post, fav.PostID)
		result = append(result, gin.H{
			"post_id":    post.ID,
			"content":    post.Content,
			"media_url":  post.MediaURL,
			"created_at": fav.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(result),
		"list":  result,
	})
}

// GetPostFavorites 获取动态的收藏用户列表（由动态作者决定是否显示）
func GetPostFavorites(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的动态ID",
		})
		return
	}

	var post model.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "动态不存在",
		})
		return
	}

	// 获取当前用户
	rawUserID, exists := c.Get("current_user_id")
	var currentUserID uint
	if exists {
		if uid, ok := rawUserID.(uint); ok {
			currentUserID = uid
		}
	}

	// 作者自己永远可见，其他人需要检查作者设置
	if currentUserID != post.UserID && !post.ShowFavorites {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "作者未公开收藏列表",
		})
		return
	}

	var favorites []model.Favorite
	config.DB.Where("post_id = ?", postID).Find(&favorites)

	var result []gin.H
	for _, fav := range favorites {
		var user model.User
		config.DB.First(&user, fav.UserID)
		result = append(result, gin.H{
			"user_id":      user.ID,
			"username":     user.Name,
			"collected_at": fav.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(result),
		"list":  result,
	})
}
