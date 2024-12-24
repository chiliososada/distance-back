package handler

import (
	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Follow 关注用户
// @Summary 关注用户
// @Description 关注指定用户
// @Tags 用户关系
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "目标用户UUID"
// @Success 200 {object} response.Response "关注成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/relationships/{id}/follow [post]
func (h *Handler) Follow(c *gin.Context) {
	// 获取当前用户UID
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	// 获取目标用户UID
	targetUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	// 不能关注自己
	if userUID == targetUID {
		Error(c, errors.ErrInvalidFollowing)
		return
	}

	if err := h.relationshipService.Follow(c.Request.Context(), userUID, targetUID); err != nil {
		logger.Error("关注用户失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID),
			logger.String("target_uid", targetUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "目标用户UUID"
// @Success 200 {object} response.Response "取消成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/relationships/{id}/follow [delete]
func (h *Handler) Unfollow(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	targetUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.relationshipService.Unfollow(c.Request.Context(), userUID, targetUID); err != nil {
		logger.Error("取消关注失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID),
			logger.String("target_uid", targetUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "关注者UUID"
// @Success 200 {object} response.Response "接受成功"
// @Failure 400,401,404 {object} response.Response "错误详情"
// @Router /api/v1/relationships/{id}/accept [post]
func (h *Handler) AcceptFollow(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	followerUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.relationshipService.AcceptFollow(c.Request.Context(), userUID, followerUID); err != nil {
		logger.Error("接受关注请求失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID),
			logger.String("follower_uid", followerUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "关注者UUID"
// @Success 200 {object} response.Response "拒绝成功"
// @Failure 400,401,404 {object} response.Response "错误详情"
// @Router /api/v1/relationships/{id}/reject [post]
func (h *Handler) RejectFollow(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	followerUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.relationshipService.RejectFollow(c.Request.Context(), userUID, followerUID); err != nil {
		logger.Error("拒绝关注请求失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID),
			logger.String("follower_uid", followerUID))
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetFollowers 获取粉丝列表
// @Summary 获取粉丝列表
// @Description 分页获取用户的粉丝列表
// @Tags 用户关系
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param status query string false "关系状态(pending/accepted/blocked)"
// @Param page query int true "页码" minimum(1)
// @Param size query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=response.RelationshipListResponse} "粉丝列表"
// @Failure 400,401 {object} response.Response "错误详情"
// @Router /api/v1/relationships/followers [get]
func (h *Handler) GetFollowers(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
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
		userUID,
		req.Status,
		req.Page,
		req.Size,
	)
	if err != nil {
		logger.Error("获取粉丝列表失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID))
		Error(c, err)
		return
	}

	Success(c, response.ToRelationshipListResponse(followers, total, req.Page, req.Size))
}

// GetFollowings 获取关注列表
// @Summary 获取关注列表
// @Description 分页获取用户的关注列表
// @Tags 用户关系
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param status query string false "关系状态(pending/accepted/blocked)"
// @Param page query int true "页码" minimum(1)
// @Param size query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=response.RelationshipListResponse} "关注列表"
// @Failure 400,401 {object} response.Response "错误详情"
// @Router /api/v1/relationships/followings [get]
func (h *Handler) GetFollowings(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
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
		userUID,
		req.Status,
		req.Page,
		req.Size,
	)
	if err != nil {
		logger.Error("获取关注列表失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID))
		Error(c, err)
		return
	}

	Success(c, response.ToRelationshipListResponse(followings, total, req.Page, req.Size))
}

// CheckRelationship 检查与指定用户的关系状态
// @Summary 检查关系状态
// @Description 检查与指定用户的关系状态(关注、被关注、好友、拉黑等)
// @Tags 用户关系
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "目标用户UUID"
// @Success 200 {object} response.Response{data=response.RelationshipStatusResponse} "关系状态"
// @Failure 400,401 {object} response.Response "错误详情"
// @Router /api/v1/relationships/{id} [get]
func (h *Handler) CheckRelationship(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	targetUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	isFollowing, _ := h.relationshipService.IsFollowing(c.Request.Context(), userUID, targetUID)
	isFollowed, _ := h.relationshipService.IsFollowed(c.Request.Context(), userUID, targetUID)
	isFriend, _ := h.relationshipService.IsFriend(c.Request.Context(), userUID, targetUID)
	IsRejected, _ := h.relationshipService.IsRejected(c.Request.Context(), targetUID, userUID)

	status := &response.RelationshipStatusResponse{
		IsFollowing: isFollowing,
		IsFollowed:  isFollowed,
		IsFriend:    isFriend,
		IsRejected:  IsRejected,
	}

	Success(c, status)
}

// GetFriends 获取好友列表
// @Summary 获取好友列表
// @Description 分页获取用户的好友列表(互相关注的用户)
// @Tags 用户关系
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param page query int true "页码" minimum(1)
// @Param size query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=response.FriendListResponse} "好友列表"
// @Failure 400,401 {object} response.Response "错误详情"
// @Router /api/v1/relationships/friends [get]
func (h *Handler) GetFriends(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	page, size, err := GetPagination(c)
	if err != nil {
		Error(c, err)
		return
	}

	friends, total, err := h.relationshipService.GetFriends(
		c.Request.Context(),
		userUID,
		page,
		size,
	)
	if err != nil {
		logger.Error("获取好友列表失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID))
		Error(c, err)
		return
	}
	Success(c, response.ToFriendListResponse(friends, nil, nil, total, page, size))
}
