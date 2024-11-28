package handler

import (
	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/gin-gonic/gin"
)

// Follow 关注用户
// @Summary 关注用户
// @Description 关注指定用户
// @Tags 用户关系
// @Accept json
// @Produce json
// @Param request body request.FollowRequest true "关注请求"
// @Success 200 {object} response.Response
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/relationships/follow [post]
func (h *Handler) Follow(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	var req request.FollowRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	// 不能关注自己
	if userID == req.TargetID {
		Error(c, errors.ErrInvalidFollowing)
		return
	}

	if err := h.relationshipService.Follow(c.Request.Context(), userID, req.TargetID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// Unfollow 取消关注
// @Summary 取消关注
// @Description 取消关注指定用户
// @Tags 用户关系
// @Accept json
// @Produce json
// @Param target_id path uint64 true "目标用户ID"
// @Success 200 {object} response.Response
// @Failure 400,401 {object} response.Response
// @Router /api/v1/relationships/following/{target_id} [delete]
func (h *Handler) Unfollow(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	targetID, err := ParseUint64Param(c, "target_id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.relationshipService.Unfollow(c.Request.Context(), userID, targetID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// AcceptFollow 接受关注请求
// @Summary 接受关注请求
// @Description 接受指定用户的关注请求
// @Tags 用户关系
// @Accept json
// @Produce json
// @Param follower_id path uint64 true "关注者ID"
// @Success 200 {object} response.Response
// @Failure 400,401,404 {object} response.Response
// @Router /api/v1/relationships/followers/{follower_id}/accept [post]
func (h *Handler) AcceptFollow(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	followerID, err := ParseUint64Param(c, "follower_id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.relationshipService.AcceptFollow(c.Request.Context(), userID, followerID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RejectFollow 拒绝关注请求
// @Summary 拒绝关注请求
// @Description 拒绝指定用户的关注请求
// @Tags 用户关系
// @Accept json
// @Produce json
// @Param follower_id path uint64 true "关注者ID"
// @Success 200 {object} response.Response
// @Failure 400,401,404 {object} response.Response
// @Router /api/v1/relationships/followers/{follower_id}/reject [post]
func (h *Handler) RejectFollow(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	followerID, err := ParseUint64Param(c, "follower_id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.relationshipService.RejectFollow(c.Request.Context(), userID, followerID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetFollowers 获取粉丝列表
// @Summary 获取粉丝列表
// @Description 获取当前用户的粉丝列表
// @Tags 用户关系
// @Produce json
// @Param request query request.GetRelationshipsRequest true "查询参数"
// @Success 200 {object} response.Response{data=response.RelationshipListResponse}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/relationships/followers [get]
func (h *Handler) GetFollowers(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	var req request.GetRelationshipsRequest
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	followers, total, err := h.relationshipService.GetFollowers(
		c.Request.Context(),
		userID,
		req.Status,
		req.Page,
		req.Size,
	)
	if err != nil {
		Error(c, err)
		return
	}

	// 转换响应
	items := make([]*response.RelationshipResponse, len(followers))
	for i, follower := range followers {
		items[i] = response.ToRelationshipResponse(follower)
	}

	Success(c, response.NewPaginatedResponse(items, total, req.Page, req.Size))
}

// GetFollowings 获取关注列表
// @Summary 获取关注列表
// @Description 获取当前用户的关注列表
// @Tags 用户关系
// @Produce json
// @Param request query request.GetRelationshipsRequest true "查询参数"
// @Success 200 {object} response.Response{data=response.RelationshipListResponse}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/relationships/following [get]
func (h *Handler) GetFollowings(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	var req request.GetRelationshipsRequest
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	followings, total, err := h.relationshipService.GetFollowings(
		c.Request.Context(),
		userID,
		req.Status,
		req.Page,
		req.Size,
	)
	if err != nil {
		Error(c, err)
		return
	}

	// 转换响应
	items := make([]*response.RelationshipResponse, len(followings))
	for i, following := range followings {
		items[i] = response.ToRelationshipResponse(following)
	}

	Success(c, response.NewPaginatedResponse(items, total, req.Page, req.Size))
}

// CheckRelationship 检查与指定用户的关系状态
// @Summary 检查关系状态
// @Description 检查与指定用户的关系状态
// @Tags 用户关系
// @Produce json
// @Param target_id path uint64 true "目标用户ID"
// @Success 200 {object} response.Response{data=response.RelationshipStatusResponse}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/relationships/status/{target_id} [get]
func (h *Handler) CheckRelationship(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	targetID, err := ParseUint64Param(c, "target_id")
	if err != nil {
		Error(c, err)
		return
	}

	// 获取关系状态
	isFollowing, _ := h.relationshipService.IsFollowing(c.Request.Context(), userID, targetID)
	isFollowed, _ := h.relationshipService.IsFollowed(c.Request.Context(), userID, targetID)
	isFriend, _ := h.relationshipService.IsFriend(c.Request.Context(), userID, targetID)
	isBlocked, _ := h.relationshipService.IsBlocked(c.Request.Context(), targetID, userID)

	status := &response.RelationshipStatusResponse{
		IsFollowing: isFollowing,
		IsFollowed:  isFollowed,
		IsFriend:    isFriend,
		IsBlocked:   isBlocked,
	}

	Success(c, status)
}

// GetFriends 获取好友列表
// @Summary 获取好友列表
// @Description 获取当前用户的好友列表（互相关注的用户）
// @Tags 用户关系
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param size query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=response.PaginatedResponse{list=[]response.UserInfo}}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/relationships/friends [get]
func (h *Handler) GetFriends(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	var req request.PaginationQuery
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	friends, total, err := h.relationshipService.GetFriends(
		c.Request.Context(),
		userID,
		req.Page,
		req.Size,
	)
	if err != nil {
		Error(c, err)
		return
	}

	// 转换响应
	friendResponses := make([]*response.UserInfo, len(friends))
	for i, friend := range friends {
		friendResponses[i] = response.ToUserInfo(friend)
	}

	Success(c, response.NewPaginatedResponse(friendResponses, total, req.Page, req.Size))
}
