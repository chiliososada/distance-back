package model

import (
	"time"
)

// UserRelationship 用户关系模型
type UserRelationship struct {
	BaseModel
	FollowerID  uint64     `gorm:"uniqueIndex:unique_relationship" json:"follower_id"`
	FollowingID uint64     `gorm:"uniqueIndex:unique_relationship" json:"following_id"`
	Status      string     `gorm:"type:enum('pending','accepted','blocked');default:'pending'" json:"status"`
	AcceptedAt  *time.Time `json:"accepted_at"`
	Follower    User       `gorm:"foreignKey:FollowerID" json:"follower"`
	Following   User       `gorm:"foreignKey:FollowingID" json:"following"`
}
