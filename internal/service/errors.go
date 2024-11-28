package service

import (
	"fmt"
	"net/http"
)

// Error 服务层错误类型
type Error struct {
	Code       int         `json:"code"`              // 错误码
	Message    string      `json:"message"`           // 错误信息
	Details    interface{} `json:"details,omitempty"` // 错误详情
	HTTPStatus int         `json:"-"`                 // HTTP状态码
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return e.Message
}

// NewError 创建新的错误
func NewError(code int, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusOK,
	}
}

// WithDetails 添加错误详情
func (e *Error) WithDetails(details interface{}) *Error {
	e.Details = details
	return e
}

// WithStatus 设置HTTP状态码
func (e *Error) WithStatus(status int) *Error {
	e.HTTPStatus = status
	return e
}

// 错误码定义
const (
	// 通用错误码 (1xxxx)
	CodeInvalidRequest   = 10001
	CodeNotFound         = 10002
	CodeUnauthorized     = 10003
	CodeForbidden        = 10004
	CodeConflict         = 10005
	CodeOperationFailed  = 10006
	CodeInvalidOperation = 10007

	// 用户相关错误码 (2xxxx)
	CodeUserNotFound        = 20001
	CodeUserAlreadyExists   = 20002
	CodeInvalidCredentials  = 20003
	CodeInvalidUserStatus   = 20004
	CodeInvalidPrivacyLevel = 20005
	CodeBlockedUser         = 20006

	// 关系相关错误码 (3xxxx)
	CodeSelfRelation        = 30001
	CodeAlreadyFollowing    = 30002
	CodeNotFollowing        = 30003
	CodeInvalidRelationType = 30004

	// 话题相关错误码 (4xxxx)
	CodeTopicNotFound      = 40001
	CodeInvalidTopicStatus = 40002
	CodeTopicExpired       = 40003
	CodeInvalidInteraction = 40004

	// 聊天相关错误码 (5xxxx)
	CodeChatRoomNotFound   = 50001
	CodeNotRoomMember      = 50002
	CodeInvalidMessageType = 50003
	CodeMessageNotFound    = 50004

	// 标签相关错误码 (6xxxx)
	CodeTagNotFound    = 60001
	CodeTooManyTags    = 60002
	CodeInvalidTagName = 60003
	CodeTagExists      = 60004

	// 文件相关错误码 (7xxxx)
	CodeInvalidFile             = 70001
	CodeInvalidFileType         = 70002
	CodeUploadFailed            = 70003
	CodeDeleteFailed            = 70004
	CodeFileNotFound            = 70005
	CodeFileTypeNotSupported    = 70006
	CodeFileTooLarge            = 70007
	CodeImageDimensionsTooLarge = 70008

	// 位置相关错误码 (8xxxx)
	CodeInvalidLocation  = 80001
	CodeLocationDisabled = 80002
	CodeLocationNotFound = 80003
)

// 预定义错误
var (
	// 通用错误
	ErrInvalidRequest = NewError(CodeInvalidRequest, "invalid request").
				WithStatus(http.StatusBadRequest)
	ErrNotFound = NewError(CodeNotFound, "resource not found").
			WithStatus(http.StatusNotFound)
	ErrUnauthorized = NewError(CodeUnauthorized, "unauthorized").
			WithStatus(http.StatusUnauthorized)
	ErrForbidden = NewError(CodeForbidden, "forbidden").
			WithStatus(http.StatusForbidden)
	ErrConflict = NewError(CodeConflict, "resource already exists").
			WithStatus(http.StatusConflict)
	ErrOperationFailed = NewError(CodeOperationFailed, "operation failed").
				WithStatus(http.StatusInternalServerError)

	// 用户相关错误
	ErrUserNotFound = NewError(CodeUserNotFound, "user not found").
			WithStatus(http.StatusNotFound)
	ErrUserAlreadyExists = NewError(CodeUserAlreadyExists, "user already exists").
				WithStatus(http.StatusConflict)
	ErrInvalidCredentials = NewError(CodeInvalidCredentials, "invalid credentials").
				WithStatus(http.StatusUnauthorized)
	ErrInvalidUserStatus = NewError(CodeInvalidUserStatus, "invalid user status").
				WithStatus(http.StatusBadRequest)
	ErrInvalidPrivacyLevel = NewError(CodeInvalidPrivacyLevel, "invalid privacy level").
				WithStatus(http.StatusBadRequest)
	ErrBlockedUser = NewError(CodeBlockedUser, "user is blocked").
			WithStatus(http.StatusForbidden)

	// 关系相关错误
	ErrSelfRelation = NewError(CodeSelfRelation, "cannot follow/block yourself").
			WithStatus(http.StatusBadRequest)
	ErrAlreadyFollowing = NewError(CodeAlreadyFollowing, "already following this user").
				WithStatus(http.StatusConflict)
	ErrNotFollowing = NewError(CodeNotFollowing, "not following this user").
			WithStatus(http.StatusBadRequest)
	ErrInvalidRelationType = NewError(CodeInvalidRelationType, "invalid relation type").
				WithStatus(http.StatusBadRequest)

	// 话题相关错误
	ErrTopicNotFound = NewError(CodeTopicNotFound, "topic not found").
				WithStatus(http.StatusNotFound)
	ErrInvalidTopicStatus = NewError(CodeInvalidTopicStatus, "invalid topic status").
				WithStatus(http.StatusBadRequest)
	ErrTopicExpired = NewError(CodeTopicExpired, "topic has expired").
			WithStatus(http.StatusBadRequest)
	ErrInvalidInteraction = NewError(CodeInvalidInteraction, "invalid interaction type").
				WithStatus(http.StatusBadRequest)

	// 聊天相关错误
	ErrChatRoomNotFound = NewError(CodeChatRoomNotFound, "chat room not found").
				WithStatus(http.StatusNotFound)
	ErrNotRoomMember = NewError(CodeNotRoomMember, "not a member of this chat room").
				WithStatus(http.StatusForbidden)
	ErrInvalidMessageType = NewError(CodeInvalidMessageType, "invalid message type").
				WithStatus(http.StatusBadRequest)
	ErrMessageNotFound = NewError(CodeMessageNotFound, "message not found").
				WithStatus(http.StatusNotFound)

	// 标签相关错误
	ErrTagNotFound = NewError(CodeTagNotFound, "tag not found").
			WithStatus(http.StatusNotFound)
	ErrTooManyTags = NewError(CodeTooManyTags, "too many tags").
			WithStatus(http.StatusBadRequest)
	ErrInvalidTagName = NewError(CodeInvalidTagName, "invalid tag name").
				WithStatus(http.StatusBadRequest)
	ErrTagExists = NewError(CodeTagExists, "tag already exists").
			WithStatus(http.StatusConflict)

	// 文件相关错误
	ErrInvalidFile = NewError(CodeInvalidFile, "invalid file").
			WithStatus(http.StatusBadRequest)
	ErrInvalidFileType = NewError(CodeInvalidFileType, "invalid file type").
				WithStatus(http.StatusBadRequest)
	ErrUploadFailed = NewError(CodeUploadFailed, "failed to upload file").
			WithStatus(http.StatusInternalServerError)
	ErrDeleteFailed = NewError(CodeDeleteFailed, "failed to delete file").
			WithStatus(http.StatusInternalServerError)
	ErrFileNotFound = NewError(CodeFileNotFound, "file not found").
			WithStatus(http.StatusNotFound)
	ErrFileTypeNotSupported = NewError(CodeFileTypeNotSupported, "file type not supported").
				WithStatus(http.StatusBadRequest)
	ErrFileTooLarge = NewError(CodeFileTooLarge, "file too large").
			WithStatus(http.StatusBadRequest)
	ErrImageDimensionsTooLarge = NewError(CodeImageDimensionsTooLarge, "image dimensions too large").
					WithStatus(http.StatusBadRequest)

	// 位置相关错误
	ErrInvalidLocation = NewError(CodeInvalidLocation, "invalid location coordinates").
				WithStatus(http.StatusBadRequest)
	ErrLocationDisabled = NewError(CodeLocationDisabled, "location sharing is disabled").
				WithStatus(http.StatusForbidden)
	ErrLocationNotFound = NewError(CodeLocationNotFound, "location not found").
				WithStatus(http.StatusNotFound)

	// 业务相关错误
	ErrInvalidStatus = NewError(CodeInvalidOperation, "invalid status").
				WithStatus(http.StatusBadRequest)
	ErrInvalidOperation = NewError(CodeInvalidOperation, "invalid operation").
				WithStatus(http.StatusBadRequest)
)

// WrapError 包装错误并添加详情
func WrapError(err error, details interface{}) *Error {
	if e, ok := err.(*Error); ok {
		return e.WithDetails(details)
	}
	return ErrOperationFailed.WithDetails(fmt.Sprintf("%v", err))
}
