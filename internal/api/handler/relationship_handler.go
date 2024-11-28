package handler

import (
	"github.com/chiliososada/distance-back/internal/service"

	"github.com/gin-gonic/gin"
)

// Follow 关注用户
func (h *Handler) Follow(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	targetID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 不能关注自己
	if userID == targetID {
		Error(c, service.ErrSelfRelation)
		return
	}

	if err := h.relationshipService.Follow(c, userID, targetID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// Unfollow 取消关注
func (h *Handler) Unfollow(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	targetID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.relationshipService.Unfollow(c, userID, targetID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// AcceptFollow 接受关注请求
func (h *Handler) AcceptFollow(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	followerID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.relationshipService.AcceptFollow(c, userID, followerID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RejectFollow 拒绝关注请求
func (h *Handler) RejectFollow(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	followerID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.relationshipService.RejectFollow(c, userID, followerID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetFollowers 获取粉丝列表
func (h *Handler) GetFollowers(c *gin.Context) {
	targetID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var query struct {
		PaginationQuery
		Status string `form:"status" binding:"omitempty,oneof=pending accepted"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	followers, total, err := h.relationshipService.GetFollowers(c, targetID, query.Status, query.Page, query.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"followers": followers,
		"total":     total,
		"page":      query.Page,
		"size":      query.PageSize,
	})
}

// GetFollowings 获取关注列表
func (h *Handler) GetFollowings(c *gin.Context) {
	targetID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var query struct {
		PaginationQuery
		Status string `form:"status" binding:"omitempty,oneof=pending accepted"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	followings, total, err := h.relationshipService.GetFollowings(c, targetID, query.Status, query.Page, query.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"followings": followings,
		"total":      total,
		"page":       query.Page,
		"size":       query.PageSize,
	})
}

// GetFriends 获取好友列表
func (h *Handler) GetFriends(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	var query PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	friends, total, err := h.relationshipService.GetFriends(c, userID, query.Page, query.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"friends": friends,
		"total":   total,
		"page":    query.Page,
		"size":    query.PageSize,
	})
}

// CheckRelationship 检查与指定用户的关系
func (h *Handler) CheckRelationship(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	targetID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 获取各种关系状态
	isFollowing, err := h.relationshipService.IsFollowing(c, userID, targetID)
	if err != nil {
		Error(c, err)
		return
	}

	isFollowed, err := h.relationshipService.IsFollowed(c, userID, targetID)
	if err != nil {
		Error(c, err)
		return
	}

	isBlocked, err := h.relationshipService.IsBlocked(c, targetID, userID)
	if err != nil {
		Error(c, err)
		return
	}

	isFriend, err := h.relationshipService.IsFriend(c, userID, targetID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"is_following": isFollowing,
		"is_followed":  isFollowed,
		"is_blocked":   isBlocked,
		"is_friend":    isFriend,
	})
}
