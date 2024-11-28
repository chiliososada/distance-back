package handler

import (
	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateGroupRequest 创建群聊请求
type CreateGroupRequest struct {
	Name           string   `json:"name" binding:"required,min=1,max=100"`
	InitialMembers []uint64 `json:"initial_members" binding:"required,min=1,dive,min=1"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ContentType string `json:"content_type" binding:"required,oneof=text image file system"`
	Content     string `json:"content" binding:"required"`
}

// CreatePrivateRoom 创建私聊
func (h *Handler) CreatePrivateRoom(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	targetID, err := ParseUint64Param(c, "target_id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 检查是否被拉黑
	isBlocked, err := h.relationshipService.IsBlocked(c, targetID, userID)
	if err != nil {
		Error(c, err)
		return
	}
	if isBlocked {
		Error(c, service.ErrBlockedUser)
		return
	}

	room, err := h.chatService.CreatePrivateRoom(c, userID, targetID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, room)
}

// CreateGroupRoom 创建群聊
func (h *Handler) CreateGroupRoom(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	room, err := h.chatService.CreateGroupRoom(c, userID, req.Name, req.InitialMembers)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, room)
}

// SendMessage 发送消息
func (h *Handler) SendMessage(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 处理媒体文件
	var files []*model.File
	if req.ContentType == "image" || req.ContentType == "file" {
		form, err := c.MultipartForm()
		if err != nil {
			Error(c, service.ErrInvalidRequest)
			return
		}

		uploadedFiles := form.File["files"]
		if len(uploadedFiles) == 0 {
			Error(c, service.ErrInvalidRequest)
			return
		}

		files = make([]*model.File, 0, len(uploadedFiles))
		for _, file := range uploadedFiles {
			files = append(files, &model.File{
				File: file,
				Type: req.ContentType,
				Name: file.Filename,
				Size: uint(file.Size),
			})
		}
	}

	message, err := h.chatService.SendMessage(c, userID, roomID, req.ContentType, req.Content, files)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, message)
}

// GetMessages 获取消息历史
func (h *Handler) GetMessages(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var query struct {
		BeforeID uint64 `form:"before_id"`
		Limit    int    `form:"limit,default=20" binding:"min=1,max=50"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	messages, err := h.chatService.GetMessages(c, userID, roomID, query.BeforeID, query.Limit)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, messages)
}

// MarkMessagesAsRead 标记消息为已读
func (h *Handler) MarkMessagesAsRead(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var req struct {
		MessageID uint64 `json:"message_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.chatService.MarkMessagesAsRead(c, userID, roomID, req.MessageID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// ListRooms 获取聊天室列表
func (h *Handler) ListRooms(c *gin.Context) {
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

	rooms, total, err := h.chatService.ListUserRooms(c, userID, query.Page, query.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"rooms": rooms,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

// GetRoomInfo 获取聊天室信息
func (h *Handler) GetRoomInfo(c *gin.Context) {
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	room, err := h.chatService.GetRoomInfo(c, roomID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, room)
}

// AddMember 添加成员
func (h *Handler) AddMember(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.chatService.AddMember(c, userID, roomID, req.UserID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RemoveMember 移除成员
func (h *Handler) RemoveMember(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	memberID, err := ParseUint64Param(c, "member_id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.chatService.RemoveMember(c, userID, roomID, memberID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// UpdateMemberRole 更新成员角色
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	memberID, err := ParseUint64Param(c, "member_id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=owner admin member"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.chatService.UpdateMemberRole(c, userID, roomID, memberID, req.Role); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// PinRoom 置顶聊天室
func (h *Handler) PinRoom(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.chatService.PinRoom(c, userID, roomID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// UnpinRoom 取消置顶聊天室
func (h *Handler) UnpinRoom(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.chatService.UnpinRoom(c, userID, roomID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetUnreadCount 获取未读消息数
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	count, err := h.chatService.GetUnreadCount(c, userID, roomID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"unread_count": count})
}
