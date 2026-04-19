package model

import "time"

type Message struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FromUserID  uint      `gorm:"index;not null" json:"from_user_id"`
	ToUserID    uint      `gorm:"index;not null" json:"to_user_id"`
	Type        string    `gorm:"size:20;not null" json:"type"`
	Content     string    `gorm:"type:text" json:"content"`
	SharePostID uint      `json:"share_post_id"`
	IsRead      bool      `gorm:"default:false" json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Message) TableName() string {
	return "messages"
}

type MessageConversation struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserA      uint      `gorm:"index;not null" json:"user_a"`
	UserB      uint      `gorm:"index;not null" json:"user_b"`
	ARemaining int       `gorm:"default:0" json:"a_remaining"`
	BRemaining int       `gorm:"default:0" json:"b_remaining"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (MessageConversation) TableName() string {
	return "message_conversations"
}
