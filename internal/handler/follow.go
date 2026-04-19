package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Follow 关注用户
func Follow(c *gin.Context) {
	rawUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "请先登录",
		})
		return
	}
	followerID, ok := rawUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户ID类型错误",
		})
		return
	}

	followeeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	if followerID == uint(followeeID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不能关注自己",
		})
		return
	}

	var user model.User
	if err := config.DB.First(&user, followeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	var existing model.Follow
	if err := config.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&existing).Error; err == nil {
		if existing.Status == 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "已经关注过了",
			})
			return
		}
		config.DB.Model(&existing).Update("status", 1)
		c.JSON(http.StatusOK, gin.H{
			"message": "关注成功",
		})
		return
	}

	follow := model.Follow{
		FollowerID: followerID,
		FolloweeID: uint(followeeID),
		Status:     1,
	}
	if err := config.DB.Create(&follow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "关注失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "关注成功",
	})
}

// Unfollow 取消关注
func Unfollow(c *gin.Context) {
	rawUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "请先登录",
		})
		return
	}
	followerID, ok := rawUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户ID类型错误",
		})
		return
	}

	followeeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	result := config.DB.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Update("status", 0)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "还没有关注过",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "取消关注成功",
	})
}

// GetFollowers 获取粉丝列表
func GetFollowers(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取当前登录用户
	rawCurrentUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	currentUserID, ok := rawCurrentUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 获取目标用户信息
	var targetUser model.User
	if err := config.DB.First(&targetUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 调试输出
	fmt.Println("=== GetFollowers 调试 ===")
	fmt.Println("targetUserID:", userID)
	fmt.Println("currentUserID:", currentUserID)
	fmt.Println("targetUser.FollowVisible:", targetUser.FollowVisible)

	// 隐私检查：如果不是查看自己，且对方设置了不公开
	if currentUserID != uint(userID) && !targetUser.FollowVisible {
		c.JSON(http.StatusForbidden, gin.H{"error": "该用户未公开关注/粉丝列表"})
		return
	}

	// 查询粉丝列表
	var follows []model.Follow
	config.DB.Where("followee_id = ? AND status = 1", userID).Find(&follows)

	var result []gin.H
	for _, f := range follows {
		var user model.User
		config.DB.First(&user, f.FollowerID)
		result = append(result, gin.H{
			"user_id":     user.ID,
			"username":    user.Name,
			"avatar":      user.Avatar,
			"followed_at": f.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(result),
		"list":  result,
	})
}

// GetFollowings 获取关注列表
func GetFollowings(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	rawCurrentUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	currentUserID, ok := rawCurrentUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	var targetUser model.User
	if err := config.DB.First(&targetUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if currentUserID != uint(userID) && !targetUser.FollowVisible {
		c.JSON(http.StatusForbidden, gin.H{"error": "该用户未公开关注/粉丝列表"})
		return
	}

	var follows []model.Follow
	config.DB.Where("follower_id = ? AND status = 1", userID).Find(&follows)

	var result []gin.H
	for _, f := range follows {
		var user model.User
		config.DB.First(&user, f.FolloweeID)
		result = append(result, gin.H{
			"user_id":     user.ID,
			"username":    user.Name,
			"avatar":      user.Avatar,
			"followed_at": f.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(result),
		"list":  result,
	})
}

// IsMutualFollow 检查是否互相关注
func IsMutualFollow(c *gin.Context) {
	rawUserID, _ := c.Get("current_user_id")
	currentUserID, _ := rawUserID.(uint)

	targetUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	var count int64
	config.DB.Model(&model.Follow{}).
		Where("(follower_id = ? AND followee_id = ?) OR (follower_id = ? AND followee_id = ?)",
			currentUserID, targetUserID, targetUserID, currentUserID).
		Where("status = 1").
		Count(&count)

	c.JSON(http.StatusOK, gin.H{"is_mutual": count == 2})
}
