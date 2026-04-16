package model

import "time"

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	PostID    uint      `gorm:"index;not null" json:"post_id"`
	Content   string    `gorm:"size:500;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
