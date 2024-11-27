package service

import "errors"

var (
	// 通用错误
	ErrInvalidRequest = errors.New("invalid request")
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("resource already exists")

	// 用户相关错误
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidUserStatus   = errors.New("invalid user status")
	ErrInvalidPrivacyLevel = errors.New("invalid privacy level")
	ErrBlockedUser         = errors.New("user is blocked")

	// 关系相关错误
	ErrSelfRelation        = errors.New("cannot follow/block yourself")
	ErrAlreadyFollowing    = errors.New("already following this user")
	ErrNotFollowing        = errors.New("not following this user")
	ErrInvalidRelationType = errors.New("invalid relation type")

	// 话题相关错误
	ErrTopicNotFound      = errors.New("topic not found")
	ErrInvalidTopicStatus = errors.New("invalid topic status")
	ErrTopicExpired       = errors.New("topic has expired")
	ErrInvalidInteraction = errors.New("invalid interaction type")

	// 聊天相关错误
	ErrChatRoomNotFound   = errors.New("chat room not found")
	ErrNotRoomMember      = errors.New("not a member of this chat room")
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrMessageNotFound    = errors.New("message not found")

	// 标签相关错误
	ErrTagNotFound    = errors.New("tag not found")
	ErrTooManyTags    = errors.New("too many tags")
	ErrInvalidTagName = errors.New("invalid tag name")
	ErrTagExists      = errors.New("tag already exists")

	// 文件相关错误
	ErrInvalidFile             = errors.New("invalid file")
	ErrInvalidFileType         = errors.New("invalid file type")
	ErrUploadFailed            = errors.New("failed to upload file")
	ErrDeleteFailed            = errors.New("failed to delete file")
	ErrFileNotFound            = errors.New("file not found")
	ErrFileTypeNotSupported    = errors.New("file type not supported")
	ErrFileTooLarge            = errors.New("file too large")
	ErrImageDimensionsTooLarge = errors.New("image dimensions too large")

	// 位置相关错误
	ErrInvalidLocation  = errors.New("invalid location coordinates")
	ErrLocationDisabled = errors.New("location sharing is disabled")
	ErrLocationNotFound = errors.New("location not found")

	// 业务相关错误
	ErrInvalidStatus    = errors.New("invalid status")
	ErrOperationFailed  = errors.New("operation failed")
	ErrInvalidOperation = errors.New("invalid operation")
)
