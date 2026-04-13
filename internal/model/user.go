package model

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	Avatar    string    `json:"avatar,omitempty" gorm:"size:500;default:''"`
	Bio       string    `json:"bio" gorm:"size:200;default:'用户还没有在此留下足迹哦~'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//用户数据存储

func (User) TableName() string {
	return "user"
}
