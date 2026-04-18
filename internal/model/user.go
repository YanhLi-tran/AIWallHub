package model

import "time"

type User struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"size:100;not null"`
	Email            string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password         string    `json:"-" gorm:"size:255;not null"`
	Avatar           string    `json:"avatar,omitempty" gorm:"size:500;default:''"`
	Bio              string    `json:"bio" gorm:"size:200;default:'用户还没有在此留下足迹哦~'"`
	EmailVisible     bool      `json:"email_visible" gorm:"default:false"`
	LikesVisible     bool      `gorm:"default:false" json:"likes_visible"`
	FavoritesVisible bool      `gorm:"default:false" json:"favorites_visible"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Status           int       `gorm:"default:1" json:"status"` // 1=正常, 0=已注销
}

//用户数据存储

func (User) TableName() string {
	return "user"
}
