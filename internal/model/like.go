package model

import "time"

type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	PostID    uint      `gorm:"index;not null" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}
