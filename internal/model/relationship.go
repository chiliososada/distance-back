package model

import (
	"time"
)

// UserRelationship 用户关系模型
type UserRelationship struct {
	BaseModel
	FollowerUID  string     `gorm:"type:varchar(36);uniqueIndex:unique_relationship" json:"follower_uid"`
	FollowingUID string     `gorm:"type:varchar(36);uniqueIndex:unique_relationship" json:"following_uid"`
	Status       string     `gorm:"type:enum('pending','accepted','blocked');default:'pending'" json:"status"`
	AcceptedAt   *time.Time `json:"accepted_at"`
	Follower     User       `gorm:"foreignKey:FollowerUID;references:UID" json:"follower"`
	Following    User       `gorm:"foreignKey:FollowingUID;references:UID" json:"following"`
}
