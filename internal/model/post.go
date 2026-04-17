package model

import (
	"time"
)

type Post struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	Type          string    `gorm:"size:20;not null" json:"type"`
	Content       string    `gorm:"size:1000" json:"content"`
	MediaURL      string    `gorm:"size:500" json:"media_url"` // 图片地址
	Views         int       `gorm:"default:0" json:"views"`    // 浏览次数
	Likes         int       `gorm:"default:0" json:"likes"`    // 点赞数
	Comments      int       `gorm:"default:0" json:"comments_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Favorites     int       `gorm:"default:0" json:"favorites_count"`
	ShowLikes     bool      `gorm:"default:true" json:"show_likes"` // 是否显示点赞用户列表
	ShowFavorites bool      `gorm:"default:true" json:"show_favorites"`
}
