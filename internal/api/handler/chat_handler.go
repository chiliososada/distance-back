package handler

import (
	"strconv"

	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/service"
	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/gin-gonic/gin"
)

// CreatePrivateRoom 创建私聊
// @Summary 创建私聊
// @Description 创建与指定用户的私聊
// @Tags 聊天
// @Accept json
// @Produce json
// @Param target_id path uint64 true "目标用户ID"
// @Success 200 {object} response.Response{data=response.ChatRoomResponse}
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/private/{target_id} [post]
func (h *Handler) CreatePrivateRoom(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	targetID, err := ParseUint64Param(c, "target_id")
	if err != nil {
		Error(c, err)
		return
	}

	room, err := h.chatService.CreatePrivateRoom(c.Request.Context(), currentUserID, targetID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToChatRoomResponse(room))
}

// CreateGroupRoom 创建群聊
// @Summary 创建群聊
// @Description 创建新的群聊
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body request.CreateGroupRequest true "群聊信息"
// @Success 200 {object} response.Response{data=response.ChatRoomResponse}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/chats/groups [post]
func (h *Handler) CreateGroupRoom(c *gin.Context) {
	var req request.CreateGroupRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	currentUserID := h.GetCurrentUserID(c)
	opts := service.GroupCreateOptions{
		Name:           req.Name,
		Announcement:   req.Announcement,
		InitialMembers: req.InitialMembers,
	}

	room, err := h.chatService.CreateGroupRoom(c.Request.Context(), currentUserID, opts)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToChatRoomResponse(room))
}

// SendMessage 发送消息
// @Summary 发送消息
// @Description 在指定聊天室发送消息
// @Tags 聊天
// @Accept multipart/form-data
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Param content_type formData string true "消息类型(text/image/file)"
// @Param content formData string true "消息内容"
// @Param files formData file false "文件(可多个)"
// @Success 200 {object} response.Response{data=response.MessageResponse}
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/messages [post]
func (h *Handler) SendMessage(c *gin.Context) {
	var req request.SendMessageRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	// 处理文件上传
	var files []*model.File
	form, err := c.MultipartForm()
	if err == nil && form != nil && form.File != nil {
		for _, fileHeaders := range form.File {
			for _, fileHeader := range fileHeaders {
				files = append(files, &model.File{
					File: fileHeader,
					Type: req.ContentType,
					Name: fileHeader.Filename,
					Size: uint(fileHeader.Size),
				})
			}
		}
	}

	opts := service.MessageCreateOptions{
		ContentType: req.ContentType,
		Content:     req.Content,
		Files:       files,
	}

	msg, err := h.chatService.SendMessage(c.Request.Context(), currentUserID, roomID, opts)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToMessageResponse(msg))
}

// GetMessages 获取消息历史
// @Summary 获取消息历史
// @Description 获取指定聊天室的消息历史
// @Tags 聊天
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Param before_id query uint64 false "在此消息ID之前"
// @Param limit query int false "获取数量" default(50)
// @Success 200 {object} response.Response{data=[]response.MessageResponse}
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/messages [get]
func (h *Handler) GetMessages(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	// 获取 before_id 参数
	beforeIDStr := c.Query("before_id")
	var beforeID uint64
	if beforeIDStr != "" {
		beforeID, err = strconv.ParseUint(beforeIDStr, 10, 64)
		if err != nil {
			Error(c, errors.New(errors.CodeValidation, "Invalid before_id parameter"))
			return
		}
	}

	// 获取 limit 参数，使用默认值 50
	limit := 50
	limitStr := c.Query("limit")
	if limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil || limitInt <= 0 || limitInt > 100 {
			Error(c, errors.New(errors.CodeValidation, "Invalid limit parameter (1-100)"))
			return
		}
		limit = limitInt
	}

	messages, err := h.chatService.GetMessages(c.Request.Context(), currentUserID, roomID, beforeID, limit)
	if err != nil {
		Error(c, err)
		return
	}

	var msgResponses []*response.MessageResponse
	for _, msg := range messages {
		msgResponses = append(msgResponses, response.ToMessageResponse(msg))
	}

	Success(c, msgResponses)
}

// GetRoomInfo 获取聊天室信息
// @Summary 获取聊天室信息
// @Description 获取指定聊天室的详细信息
// @Tags 聊天
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Success 200 {object} response.Response{data=response.ChatRoomResponse}
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id} [get]
func (h *Handler) GetRoomInfo(c *gin.Context) {
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	room, err := h.chatService.GetRoomInfo(c.Request.Context(), roomID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToChatRoomResponse(room))
}

// ListRooms 获取聊天室列表
// @Summary 获取聊天室列表
// @Description 获取当前用户的所有聊天室列表
// @Tags 聊天
// @Produce json
// @Param page query int true "页码"
// @Param size query int true "每页数量"
// @Success 200 {object} response.Response{data=response.ChatRoomListResponse}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/chats [get]
func (h *Handler) ListRooms(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	page, size, err := GetPagination(c)
	if err != nil {
		Error(c, err)
		return
	}

	rooms, total, err := h.chatService.ListUserRooms(c.Request.Context(), currentUserID, page, size)
	if err != nil {
		Error(c, err)
		return
	}

	roomResponses := make([]*response.ChatRoomResponse, len(rooms))
	for i, room := range rooms {
		roomResponses[i] = response.ToChatRoomResponse(room)
	}

	Success(c, response.NewPaginatedResponse(roomResponses, total, page, size))
}

// UpdateRoom 更新聊天室信息
// @Summary 更新聊天室信息
// @Description 更新聊天室的基本信息
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Param request body request.UpdateRoomRequest true "更新信息"
// @Success 200 {object} response.Response
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id} [put]
func (h *Handler) UpdateRoom(c *gin.Context) {
	var req request.UpdateRoomRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	opts := service.RoomUpdateOptions{
		Name:         req.Name,
		Announcement: req.Announcement,
	}

	if err := h.chatService.UpdateRoom(c.Request.Context(), currentUserID, roomID, opts); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// UpdateRoomAvatar 更新聊天室头像
// @Summary 更新聊天室头像
// @Description 更新群聊的头像
// @Tags 聊天
// @Accept multipart/form-data
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Param avatar formData file true "头像文件"
// @Success 200 {object} response.Response{data=response.ChatRoomResponse}
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/avatar [put]
func (h *Handler) UpdateRoomAvatar(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		Error(c, errors.New(errors.CodeValidation, "Invalid file"))
		return
	}

	avatar := &model.File{
		File: file,
		Type: "avatar",
		Name: file.Filename,
		Size: uint(file.Size),
	}

	room, err := h.chatService.UpdateRoomAvatar(c.Request.Context(), currentUserID, roomID, avatar)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToChatRoomResponse(room))
}

// AddMember 添加成员
// @Summary 添加群聊成员
// @Description 向群聊中添加新成员
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Param user_id body uint64 true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/members [post]
func (h *Handler) AddMember(c *gin.Context) {
	var req request.AddMemberRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.AddMember(c.Request.Context(), currentUserID, roomID, req.UserID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RemoveMember 移除成员
// @Summary 移除群聊成员
// @Description 从群聊中移除指定成员
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Param member_id path uint64 true "成员ID"
// @Success 200 {object} response.Response
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/members/{member_id} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	memberID, err := ParseUint64Param(c, "member_id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.RemoveMember(c.Request.Context(), currentUserID, roomID, memberID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// UpdateMember 更新成员信息
// @Summary 更新成员信息
// @Description 更新群聊成员的角色和设置
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Param member_id path uint64 true "成员ID"
// @Param request body request.UpdateMemberRequest true "更新信息"
// @Success 200 {object} response.Response
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/members/{member_id} [put]
func (h *Handler) UpdateMember(c *gin.Context) {
	// 获取请求体
	var req request.UpdateMemberRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	// 获取当前用户 ID
	currentUserID := h.GetCurrentUserID(c)

	// 获取聊天室 ID
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	// 获取要更新的成员 ID
	memberID, err := ParseUint64Param(c, "member_id")
	if err != nil {
		Error(c, err)
		return
	}

	// 转换为服务层参数结构
	opts := service.MemberUpdateOptions{
		Role:     req.Role,
		Nickname: req.Nickname,
		IsMuted:  req.IsMuted,
	}

	// 调用服务层方法
	if err := h.chatService.UpdateMember(c.Request.Context(), currentUserID, roomID, memberID, opts); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetRoomMembers 获取成员列表
// @Summary 获取聊天室成员列表
// @Description 获取指定聊天室的所有成员
// @Tags 聊天
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Param page query int true "页码"
// @Param size query int true "每页数量"
// @Success 200 {object} response.Response{data=response.ChatMemberListResponse}
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/members [get]
func (h *Handler) GetRoomMembers(c *gin.Context) {
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	page, size, err := GetPagination(c)
	if err != nil {
		Error(c, err)
		return
	}

	members, total, err := h.chatService.GetRoomMembers(c.Request.Context(), roomID, page, size)
	if err != nil {
		Error(c, err)
		return
	}

	memberResponses := make([]*response.ChatMemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = response.ToChatMemberResponse(member)
	}

	Success(c, response.NewPaginatedResponse(memberResponses, total, page, size))
}

// LeaveRoom 退出聊天室
// @Summary 退出聊天室
// @Description 退出指定的群聊
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Success 200 {object} response.Response
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/leave [post]
func (h *Handler) LeaveRoom(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.LeaveRoom(c.Request.Context(), currentUserID, roomID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// MarkMessagesAsRead 标记消息已读
// @Summary 标记消息已读
// @Description 将指定消息及之前的所有消息标记为已读
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Param message_id path uint64 true "消息ID"
// @Success 200 {object} response.Response
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/messages/{message_id}/read [post]
func (h *Handler) MarkMessagesAsRead(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	messageID, err := ParseUint64Param(c, "message_id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.MarkMessagesAsRead(c.Request.Context(), currentUserID, roomID, messageID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetUnreadCount 获取未读消息数
// @Summary 获取未读消息数
// @Description 获取指定聊天室的未读消息数量
// @Tags 聊天
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Success 200 {object} response.Response{data=map[string]uint64}
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/unread [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	count, err := h.chatService.GetUnreadCount(c.Request.Context(), currentUserID, roomID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"unread_count": count})
}

// PinRoom 置顶聊天室
// @Summary 置顶聊天室
// @Description 将指定聊天室置顶
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Success 200 {object} response.Response
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/pin [post]
func (h *Handler) PinRoom(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.PinRoom(c.Request.Context(), currentUserID, roomID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// UnpinRoom 取消置顶聊天室
// @Summary 取消置顶聊天室
// @Description 取消指定聊天室的置顶状态
// @Tags 聊天
// @Accept json
// @Produce json
// @Param id path uint64 true "聊天室ID"
// @Success 200 {object} response.Response
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/chats/{id}/pin [delete]
func (h *Handler) UnpinRoom(c *gin.Context) {
	currentUserID := h.GetCurrentUserID(c)
	roomID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.UnpinRoom(c.Request.Context(), currentUserID, roomID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// UpdateMemberRole 更新聊天室成员角色
// @Summary 更新成员角色
// @Description 更新聊天室成员的角色(仅管理员可操作)
// @Tags 聊天
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path uint64 true "聊天室ID"
// @Param member_id path uint64 true "成员ID"
// @Param request body request.UpdateMemberRoleRequest true "角色信息"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/members/{member_id}/role [put]
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	// 1. 获取当前用户ID
	currentUserID := h.GetCurrentUserID(c)
	if currentUserID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 2. 获取路径参数
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

	// 3. 解析请求体
	var req request.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.NewError(service.CodeInvalidRequest, "无效的角色参数"))
		return
	}

	// 4. 验证角色参数
	if !isValidMemberRole(req.Role) {
		Error(c, service.NewError(service.CodeInvalidRequest, "无效的角色类型"))
		return
	}

	// 5. 调用服务层更新角色
	opts := service.MemberUpdateOptions{
		Role: req.Role,
	}

	if err := h.chatService.UpdateMember(c.Request.Context(), currentUserID, roomID, memberID, opts); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// isValidMemberRole 检查成员角色是否有效
func isValidMemberRole(role string) bool {
	return role == "owner" || role == "admin" || role == "member"
}
