package handler

import (
	"AIWallHub/config"
	"AIWallHub/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SendMessage(c *gin.Context) {
	// 获取当前用户
	rawUserID, exists := c.Get("current_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	fromUserID, ok := rawUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 获取目标用户ID
	toUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 解析请求体
	var json struct {
		Type        string `json:"type"`
		Content     string `json:"content"`
		SharePostID uint   `json:"share_post_id"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	// 给自己发消息，无限制
	if fromUserID == uint(toUserID) {
		message := model.Message{
			FromUserID:  fromUserID,
			ToUserID:    uint(toUserID),
			Type:        json.Type,
			Content:     json.Content,
			SharePostID: json.SharePostID,
			IsRead:      true,
		}
		if err := config.DB.Create(&message).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "发送成功", "message_id": message.ID})
		return
	}

	// 检查是否互相关注
	var mutualCount int64
	config.DB.Model(&model.Follow{}).
		Where("(follower_id = ? AND followee_id = ?) OR (follower_id = ? AND followee_id = ?)",
			fromUserID, toUserID, toUserID, fromUserID).
		Where("status = 1").
		Count(&mutualCount)
	isMutual := mutualCount == 2

	// 互相关注，无限制
	if isMutual {
		message := model.Message{
			FromUserID:  fromUserID,
			ToUserID:    uint(toUserID),
			Type:        json.Type,
			Content:     json.Content,
			SharePostID: json.SharePostID,
			IsRead:      false,
		}
		if err := config.DB.Create(&message).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "发送成功", "message_id": message.ID})
		return
	}

	// ========== 未互相关注 ==========
	// 获取或创建对话记录（user_a < user_b）
	userA := fromUserID
	userB := uint(toUserID)
	if userA > userB {
		userA, userB = userB, userA
	}

	var conv model.MessageConversation
	config.DB.Where("user_a = ? AND user_b = ?", userA, userB).First(&conv)

	// 统计双方发送的消息数量
	var msgFromMeToOther int64
	var msgFromOtherToMe int64
	config.DB.Model(&model.Message{}).
		Where("from_user_id = ? AND to_user_id = ?", fromUserID, toUserID).
		Count(&msgFromMeToOther)
	config.DB.Model(&model.Message{}).
		Where("from_user_id = ? AND to_user_id = ?", toUserID, fromUserID).
		Count(&msgFromOtherToMe)

	// ========== 情况1：第一条消息 ==========
	if msgFromMeToOther == 0 && msgFromOtherToMe == 0 {
		message := model.Message{
			FromUserID:  fromUserID,
			ToUserID:    uint(toUserID),
			Type:        json.Type,
			Content:     json.Content,
			SharePostID: json.SharePostID,
			IsRead:      false,
		}
		if err := config.DB.Create(&message).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败"})
			return
		}

		if conv.ID == 0 {
			conv = model.MessageConversation{
				UserA:      userA,
				UserB:      userB,
				ARemaining: 0,
				BRemaining: 0,
			}
			config.DB.Create(&conv)
		}

		c.JSON(http.StatusOK, gin.H{"message": "发送成功", "message_id": message.ID})
		return
	}

	// ========== 情况2：自己发过，对方没回复 ==========
	if msgFromMeToOther > 0 && msgFromOtherToMe == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "消息额度已用完，等待对方回复"})
		return
	}

	// ========== 情况3：对方发过消息 ==========
	if msgFromOtherToMe > 0 {
		// 重新查询获取最新额度
		config.DB.Where("user_a = ? AND user_b = ?", userA, userB).First(&conv)

		// 确定当前用户的剩余次数
		var remaining int
		if fromUserID == conv.UserA {
			remaining = conv.ARemaining
		} else {
			remaining = conv.BRemaining
		}

		// 检查是否可以发送
		// 情况3a：这是对方回复后的第一次发送（remaining == 0 且 msgFromMeToOther == 0）
		// 情况3b：正常发送（remaining > 0）
		if remaining == 0 && msgFromMeToOther == 0 {
			// 这是第一次回复后的发送，可以发
		} else if remaining > 0 {
			// 正常发送
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "消息额度已用完"})
			return
		}

		// 发送消息
		message := model.Message{
			FromUserID:  fromUserID,
			ToUserID:    uint(toUserID),
			Type:        json.Type,
			Content:     json.Content,
			SharePostID: json.SharePostID,
			IsRead:      false,
		}
		if err := config.DB.Create(&message).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败"})
			return
		}

		// 更新额度
		if fromUserID == conv.UserA {
			if remaining == 0 && msgFromMeToOther == 0 {
				// 第一次发送，重置为5并消耗1
				config.DB.Model(&conv).Update("a_remaining", 4)
			} else {
				config.DB.Model(&conv).Update("a_remaining", remaining-1)
			}
			// 重置对方的额度为5
			config.DB.Model(&conv).Update("b_remaining", 5)
		} else {
			if remaining == 0 && msgFromMeToOther == 0 {
				config.DB.Model(&conv).Update("b_remaining", 4)
			} else {
				config.DB.Model(&conv).Update("b_remaining", remaining-1)
			}
			config.DB.Model(&conv).Update("a_remaining", 5)
		}

		c.JSON(http.StatusOK, gin.H{"message": "发送成功", "message_id": message.ID})
		return
	}
}

// GetMessages 获取与某人的私信列表
func GetMessages(c *gin.Context) {
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

	otherUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	offset := (page - 1) * pageSize

	var messages []model.Message
	var total int64

	config.DB.Model(&model.Message{}).
		Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
			currentUserID, otherUserID, otherUserID, currentUserID).
		Count(&total)

	config.DB.Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		currentUserID, otherUserID, otherUserID, currentUserID).
		Order("created_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages)

	// 标记已读（不标记自己的消息）
	if currentUserID != uint(otherUserID) {
		config.DB.Model(&model.Message{}).
			Where("to_user_id = ? AND from_user_id = ? AND is_read = false", currentUserID, otherUserID).
			Update("is_read", true)
	}

	var result []gin.H
	for _, msg := range messages {
		result = append(result, gin.H{
			"id":            msg.ID,
			"from_user_id":  msg.FromUserID,
			"to_user_id":    msg.ToUserID,
			"type":          msg.Type,
			"content":       msg.Content,
			"share_post_id": msg.SharePostID,
			"is_read":       msg.IsRead,
			"created_at":    msg.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"list":      result,
	})
}

// GetConversationList 获取对话列表
func GetConversationList(c *gin.Context) {
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

	// 查询所有与当前用户有关的对话
	var convs []model.MessageConversation
	config.DB.Where("user_a = ? OR user_b = ?", currentUserID, currentUserID).
		Order("updated_at DESC").
		Find(&convs)

	var result []gin.H
	for _, conv := range convs {
		var otherUserID uint
		if conv.UserA == currentUserID {
			otherUserID = conv.UserB
		} else {
			otherUserID = conv.UserA
		}

		var user model.User
		config.DB.First(&user, otherUserID)

		// 获取最后一条消息
		var lastMsg model.Message
		config.DB.Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
			currentUserID, otherUserID, otherUserID, currentUserID).
			Order("created_at DESC").
			First(&lastMsg)

		// 获取未读消息数（不包括自己的消息）
		var unreadCount int64
		if currentUserID != otherUserID {
			config.DB.Model(&model.Message{}).
				Where("from_user_id = ? AND to_user_id = ? AND is_read = false", otherUserID, currentUserID).
				Count(&unreadCount)
		}

		result = append(result, gin.H{
			"user_id":      user.ID,
			"username":     user.Name,
			"avatar":       user.Avatar,
			"last_message": lastMsg.Content,
			"last_time":    lastMsg.CreatedAt,
			"unread_count": unreadCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(result),
		"list":  result,
	})
}
