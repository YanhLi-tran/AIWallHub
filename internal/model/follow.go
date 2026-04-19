package model

import "time"

type Follow struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FollowerID uint      `gorm:"index;not null" json:"follower_id"`
	FolloweeID uint      `gorm:"index;not null" json:"followee_id"`
	Status     int       `gorm:"default:1" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Follow) TableName() string {
	return "follows"
}
