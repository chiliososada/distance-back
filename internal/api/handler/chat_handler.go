package handler

import (
	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/internal/model"

	"github.com/chiliososada/distance-back/internal/service"

	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/gin-gonic/gin"
)

// CreatePrivateRoom 创建私聊
// @Summary 创建私聊
// @Description 创建一个与指定用户的私聊房间
// @Tags 聊天
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param target_id path string true "目标用户UUID"
// @Success 200 {object} response.Response{data=response.ChatRoomResponse} "聊天室信息"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/private/{target_id} [post]
func (h *Handler) CreatePrivateRoom(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	targetUID, err := ParseUUID(c, "target_id")
	if err != nil {
		Error(c, err)
		return
	}

	if userUID == targetUID {
		Error(c, errors.ErrSelfChat)
		return
	}

	room, err := h.chatService.CreatePrivateRoom(c.Request.Context(), userUID, targetUID)
	if err != nil {
		logger.Error("创建私聊失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID),
			logger.String("target_uid", targetUID))
		Error(c, err)
		return
	}

	Success(c, response.ToChatRoomResponse(room))
}

// CreateGroupRoom 创建群聊
// @Summary 创建群聊
// @Description 创建一个新的群聊
// @Tags 聊天
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param request body request.CreateGroupRequest true "群聊信息"
// @Success 200 {object} response.Response{data=response.ChatRoomResponse} "群聊信息"
// @Failure 400,401 {object} response.Response "错误详情"
// @Router /api/v1/chats/groups [post]
func (h *Handler) CreateGroupRoom(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	var req request.CreateGroupRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	// if len(req.InitialMembers) > DefaultMaxRoomMembers {
	// 	Error(c, errors.ErrRoomMemberLimit)
	// 	return
	// }
	// 从当前上下文获取用户信息
	user, err := h.userService.GetUserByUID(c.Request.Context(), userUID)
	if err != nil {
		Error(c, err)
		return
	}
	opts := service.GroupCreateOptions{
		Name:            req.Name,
		TopicUID:        req.TopicUID,
		Announcement:    req.Announcement,
		CreatorNickname: user.Nickname, // 传递创建者昵称
	}

	room, err := h.chatService.CreateGroupRoom(c.Request.Context(), userUID, opts)
	if err != nil {
		logger.Error("创建群聊失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("creator_uid", userUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param content_type formData string true "消息类型(text/image/file/system)"
// @Param content formData string true "消息内容"
// @Param files formData file false "文件(可多个)"
// @Success 200 {object} response.Response{data=response.MessageResponse} "消息信息"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/messages [post]
func (h *Handler) SendMessage(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req request.SendMessageRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

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

	msg, err := h.chatService.SendMessage(c.Request.Context(), userUID, roomUID, opts)
	if err != nil {
		logger.Error("发送消息失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID))
		Error(c, err)
		return
	}

	Success(c, response.ToMessageResponse(msg))
}

// GetMessages 获取消息记录
// @Summary 获取消息记录
// @Description 获取指定聊天室的消息历史记录
// @Tags 聊天
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param request query request.GetMessagesRequest true "查询参数"
// @Success 200 {object} response.Response{data=response.MessageListResponse} "消息列表"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/messages [get]
func (h *Handler) GetMessages(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req request.GetMessagesRequest
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	messages, err := h.chatService.GetMessages(c.Request.Context(), userUID, roomUID, req.BeforeUID, req.Size)
	if err != nil {
		logger.Error("获取消息记录失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID))
		Error(c, err)
		return
	}

	Success(c, response.ToMessageListResponse(messages, int64(len(messages)), req.Page, req.Size))
}

// UpdateRoom 更新聊天室信息
// @Summary 更新聊天室信息
// @Description 更新群聊的基本信息
// @Tags 聊天
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param request body request.UpdateRoomRequest true "更新信息"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id} [put]
func (h *Handler) UpdateRoom(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req request.UpdateRoomRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	opts := service.RoomUpdateOptions{
		Name:         req.Name,
		Announcement: req.Announcement,
	}

	if err := h.chatService.UpdateRoom(c.Request.Context(), userUID, roomUID, opts); err != nil {
		logger.Error("更新聊天室失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID))
		Error(c, err)
		return
	}

	Success(c, nil)
}

// AddMember 添加成员
// @Summary 添加群聊成员
// @Description 向群聊中添加新成员
// @Tags 聊天
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param request body request.AddMemberRequest true "成员信息"
// @Success 200 {object} response.Response "添加成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/members [post]
func (h *Handler) AddMember(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req request.AddMemberRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.AddMember(c.Request.Context(), userUID, roomUID, req.UserUID); err != nil {
		logger.Error("添加群聊成员失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID),
			logger.String("member_uid", req.UserUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param member_id path string true "成员UUID"
// @Success 200 {object} response.Response "移除成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/members/{member_id} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	memberUID, err := ParseUUID(c, "member_id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.RemoveMember(c.Request.Context(), userUID, roomUID, memberUID); err != nil {
		logger.Error("移除群聊成员失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID),
			logger.String("member_uid", memberUID))
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetRoomMembers 获取成员列表
// @Summary 获取聊天室成员列表
// @Description 获取指定聊天室的所有成员信息
// @Tags 聊天
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param page query int true "页码" minimum(1)
// @Param size query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=response.ChatMemberListResponse} "成员列表"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/members [get]
func (h *Handler) GetRoomMembers(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	page, size, err := GetPagination(c)
	if err != nil {
		Error(c, err)
		return
	}

	members, total, err := h.chatService.GetRoomMembers(c.Request.Context(), roomUID, page, size)
	if err != nil {
		logger.Error("获取群聊成员列表失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID))
		Error(c, err)
		return
	}

	Success(c, response.ToChatMemberListResponse(members, total, page, size))
}

// MarkMessagesAsRead 标记消息已读
// @Summary 标记消息已读
// @Description 标记指定消息及之前的所有消息为已读
// @Tags 聊天
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param message_id path string true "消息UUID"
// @Success 200 {object} response.Response "标记成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/messages/{message_id}/read [post]
func (h *Handler) MarkMessagesAsRead(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	messageUID, err := ParseUUID(c, "message_id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.MarkMessagesAsRead(c.Request.Context(), userUID, roomUID, messageUID); err != nil {
		logger.Error("标记消息已读失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID),
			logger.String("message_uid", messageUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Success 200 {object} response.Response{data=map[string]int64} "未读数量"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/unread [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	count, err := h.chatService.GetUnreadCount(c.Request.Context(), userUID, roomUID)
	if err != nil {
		logger.Error("获取未读消息数失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Success 200 {object} response.Response "置顶成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/pin [post]
func (h *Handler) PinRoom(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.PinRoom(c.Request.Context(), userUID, roomUID); err != nil {
		logger.Error("置顶聊天室失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Success 200 {object} response.Response "取消置顶成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/pin [delete]
func (h *Handler) UnpinRoom(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.UnpinRoom(c.Request.Context(), userUID, roomUID); err != nil {
		logger.Error("取消置顶聊天室失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID))
		Error(c, err)
		return
	}

	Success(c, nil)
}

// UpdateMemberRole 更新成员角色
// @Summary 更新成员角色
// @Description 更新群聊成员的角色
// @Tags 聊天
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param member_id path string true "成员UUID"
// @Param request body request.UpdateMemberRoleRequest true "角色信息"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/members/{member_id}/role [put]
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	memberUID, err := ParseUUID(c, "member_id")
	if err != nil {
		Error(c, err)
		return
	}

	var req request.UpdateMemberRoleRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	opts := service.MemberUpdateOptions{
		Role: req.Role,
	}

	if err := h.chatService.UpdateMember(c.Request.Context(), userUID, roomUID, memberUID, opts); err != nil {
		logger.Error("更新成员角色失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID),
			logger.String("member_uid", memberUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param avatar formData file true "头像文件"
// @Success 200 {object} response.Response{data=response.ChatRoomResponse} "聊天室信息"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/avatar [put]
func (h *Handler) UpdateRoomAvatar(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		Error(c, errors.ErrInvalidFileUpload)
		return
	}

	avatar := &model.File{
		File: file,
		Type: "avatar",
		Name: file.Filename,
		Size: uint(file.Size),
	}

	room, err := h.chatService.UpdateRoomAvatar(c.Request.Context(), userUID, roomUID, avatar)
	if err != nil {
		logger.Error("更新聊天室头像失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID))
		Error(c, err)
		return
	}

	Success(c, response.ToChatRoomResponse(room))
}

// GetRoomInfo 获取聊天室信息
// @Summary 获取聊天室信息
// @Description 获取指定聊天室的详细信息
// @Tags 聊天
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Success 200 {object} response.Response{data=response.ChatRoomResponse} "聊天室信息"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id} [get]
func (h *Handler) GetRoomInfo(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	room, err := h.chatService.GetRoomInfo(c.Request.Context(), roomUID)
	if err != nil {
		logger.Error("获取聊天室信息失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID))
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
// @Param Authorization header string true "Bearer 用户令牌"
// @Param request query request.ListRoomsRequest true "查询参数"
// @Success 200 {object} response.Response{data=response.ChatRoomListResponse} "聊天室列表"
// @Failure 400,401 {object} response.Response "错误详情"
// @Router /api/v1/chats [get]
func (h *Handler) ListRooms(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	var req request.ListRoomsRequest
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	rooms, total, err := h.chatService.ListUserRooms(c.Request.Context(), userUID, req.Page, req.Size)
	if err != nil {
		logger.Error("获取聊天室列表失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID))
		Error(c, err)
		return
	}

	Success(c, response.ToChatRoomListResponse(rooms, total, req.Page, req.Size))
}

// UpdateMember 更新成员信息
// @Summary 更新成员信息
// @Description 更新群聊成员的基本信息
// @Tags 聊天
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Param member_id path string true "成员UUID"
// @Param request body request.UpdateMemberRequest true "更新信息"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/members/{member_id} [put]
func (h *Handler) UpdateMember(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	memberUID, err := ParseUUID(c, "member_id")
	if err != nil {
		Error(c, err)
		return
	}

	var req request.UpdateMemberRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	opts := service.MemberUpdateOptions{
		Role:     req.Role,
		Nickname: req.Nickname,
		IsMuted:  req.IsMuted,
	}

	if err := h.chatService.UpdateMember(c.Request.Context(), userUID, roomUID, memberUID, opts); err != nil {
		logger.Error("更新成员信息失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID),
			logger.String("member_uid", memberUID))
		Error(c, err)
		return
	}

	Success(c, nil)
}

// LeaveRoom 退出聊天室
// @Summary 退出聊天室
// @Description 退出指定的群聊(群主不能退出)
// @Tags 聊天
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "聊天室UUID"
// @Success 200 {object} response.Response "退出成功"
// @Failure 400,401,403 {object} response.Response "错误详情"
// @Router /api/v1/chats/{id}/leave [post]
func (h *Handler) LeaveRoom(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	roomUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.chatService.LeaveRoom(c.Request.Context(), userUID, roomUID); err != nil {
		logger.Error("退出聊天室失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("room_uid", roomUID),
			logger.String("user_uid", userUID))
		Error(c, err)
		return
	}

	Success(c, nil)
}
