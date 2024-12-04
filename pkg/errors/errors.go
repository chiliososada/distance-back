package errors

import (
	"fmt"
	"net/http"
)

// AppError 自定义错误结构
type AppError struct {
	Code       int         `json:"code"`              // 错误码
	Message    string      `json:"message"`           // 错误消息
	Details    interface{} `json:"details,omitempty"` // 错误详情
	Developer  string      `json:"-"`                 // 开发者错误信息，不返回给用户
	HTTPStatus int         `json:"-"`                 // HTTP状态码
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	return fmt.Sprintf("错误码: %d, 错误信息: %s", e.Code, e.Message)
}

// 错误码定义
const (
	// 成功
	CodeSuccess = 0

	// 系统级错误 (1xxxx)
	CodeUnknown         = 10000 // 未知错误
	CodeValidation      = 10001 // 参数验证错误
	CodeDatabase        = 10002 // 数据库错误
	CodeAuthentication  = 10003 // 认证失败
	CodeAuthorization   = 10004 // 授权失败
	CodeNotFound        = 10005 // 资源不存在
	CodeDuplicate       = 10006 // 资源已存在
	CodeThirdParty      = 10007 // 第三方服务错误
	CodeUpload          = 10008 // 上传失败
	CodeDownload        = 10009 // 下载失败
	CodeOperation       = 10010 // 操作失败
	CodeOperationFailed = 10011 // 操作执行失败

	// 用户相关错误 (2xxxx)
	CodeUserNotFound      = 20001 // 用户不存在
	CodeUserExists        = 20002 // 用户已存在
	CodePasswordInvalid   = 20003 // 密码错误
	CodeTokenInvalid      = 20004 // Token无效
	CodeTokenExpired      = 20005 // Token过期
	CodeUserBlocked       = 20006 // 用户已被封禁
	CodeUserNotVerified   = 20007 // 用户未验证
	CodeProfileIncomplete = 20008 // 用户资料不完整
	CodeDeviceNotFound    = 20009 // 设备不存在
	CodeDeviceExists      = 20010 // 设备已存在
	CodeInvalidFile       = 20011 // 无效文件

	// 社交关系错误 (3xxxx)
	CodeRelationExists   = 30001 // 关系已存在
	CodeRelationNotFound = 30002 // 关系不存在
	CodeSelfRelation     = 30003 // 不能关注自己
	CodeBlockedUser      = 30004 // 用户已被拉黑
	CodeFollowLimit      = 30005 // 关注数达到上限
	CodeInvalidRelation  = 30006 // 无效的关系

	// 聊天相关错误 (4xxxx)
	CodeChatRoomNotFound   = 40001 // 聊天室不存在
	CodeNotChatMember      = 40002 // 不是聊天室成员
	CodeMessageTooLong     = 40003 // 消息内容过长
	CodeMessageTypeInvalid = 40004 // 无效的消息类型
	CodeRoomFull           = 40005 // 聊天室已满
	CodeMuted              = 40006 // 用户被禁言
	CodeMessageNotFound    = 40007 // 消息不存在
	CodeForbidden          = 40008 // 群主不能退出群聊
	CodeSelfChat           = 40009 // 不能与自己创建私聊

	// 话题相关错误 (5xxxx)
	CodeTopicNotFound      = 50001 // 话题不存在
	CodeTopicExpired       = 50002 // 话题已过期
	CodeTagNotFound        = 50003 // 标签不存在
	CodeTopicLocked        = 50004 // 话题已锁定
	CodeInteractionInvalid = 50005 // 无效的互动类型
	CodeTooManyTags        = 50006 // 标签数量超限
	CodeContentInvalid     = 50007 // 内容不合规

	// 位置相关错误码 (8xxxx)
	CodeInvalidLocation  = 80001 // 无效的位置坐标
	CodeLocationDisabled = 80002 // 位置服务已禁用
	CodeLocationNotFound = 80003 // 位置信息不存在
)

// 创建新的错误
func New(code int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusOK, // 默认200
	}
}

// 包装已有错误
func Wrap(err error, code int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Developer:  err.Error(),
		HTTPStatus: http.StatusOK,
	}
}

// WithStatus 设置HTTP状态码
func (e *AppError) WithStatus(status int) *AppError {
	e.HTTPStatus = status
	return e
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// WithDeveloper 添加开发者信息
func (e *AppError) WithDeveloper(dev string) *AppError {
	e.Developer = dev
	return e
}

// Is 判断错误类型
func Is(err error, target *AppError) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == target.Code
	}
	return false
}

// 预定义错误实例
var (
	// 系统级错误
	ErrUnknown        = New(CodeUnknown, "未知错误")
	ErrValidation     = New(CodeValidation, "参数验证错误")
	ErrDatabase       = New(CodeDatabase, "数据库错误")
	ErrAuthentication = New(CodeAuthentication, "认证失败")
	ErrAuthorization  = New(CodeAuthorization, "权限不足")
	ErrNotFound       = New(CodeNotFound, "资源不存在")
	ErrDuplicate      = New(CodeDuplicate, "资源已存在")
	ErrThirdParty     = New(CodeThirdParty, "第三方服务错误")
	ErrOperation      = New(CodeOperation, "操作失败")

	// 用户相关错误
	ErrUserExists      = New(CodeUserExists, "用户已存在")
	ErrUserNotFound    = New(CodeUserNotFound, "用户不存在").WithStatus(http.StatusNotFound)
	ErrPasswordInvalid = New(CodePasswordInvalid, "密码错误")
	ErrTokenInvalid    = New(CodeTokenInvalid, "无效的Token")
	ErrUserBlocked     = New(CodeUserBlocked, "账号已被封禁")
	ErrUnauthorized    = New(CodeAuthentication, "未认证").WithStatus(http.StatusUnauthorized)
	ErrForbidden       = New(CodeAuthorization, "无权限访问").WithStatus(http.StatusForbidden)
	ErrInvalidRequest  = New(CodeValidation, "无效请求").WithStatus(http.StatusBadRequest)

	// 用户资料相关错误
	ErrInvalidProfile    = New(CodeProfileIncomplete, "无效的用户资料").WithStatus(http.StatusBadRequest)
	ErrInvalidBirthDate  = New(CodeValidation, "无效的出生日期格式").WithStatus(http.StatusBadRequest)
	ErrInvalidFileUpload = New(CodeInvalidFile, "无效的文件上传").WithStatus(http.StatusBadRequest)

	// 设备相关错误
	ErrDeviceTokenRequired  = New(CodeValidation, "设备令牌不能为空").WithStatus(http.StatusBadRequest)
	ErrDeviceRegisterFailed = New(CodeOperationFailed, "设备注册失败").WithStatus(http.StatusInternalServerError)

	// 用户操作相关错误
	ErrUserProfileUpdateFailed = New(CodeOperationFailed, "用户资料更新失败").WithStatus(http.StatusInternalServerError)
	ErrUserProfileFetchFailed  = New(CodeOperationFailed, "获取用户资料失败").WithStatus(http.StatusInternalServerError)
	ErrAvatarUpdateFailed      = New(CodeOperationFailed, "头像更新失败").WithStatus(http.StatusInternalServerError)
	ErrSearchFailed            = New(CodeOperationFailed, "用户搜索失败").WithStatus(http.StatusInternalServerError)
	ErrUserListFailed          = New(CodeOperationFailed, "获取用户列表失败").WithStatus(http.StatusInternalServerError)
	ErrNearbyUsersFailed       = New(CodeOperationFailed, "获取附近用户失败").WithStatus(http.StatusInternalServerError)

	// 隐私相关错误
	ErrPrivacyAccessDenied = New(CodeForbidden, "因隐私设置拒绝访问").WithStatus(http.StatusForbidden)

	// 社交关系错误
	ErrRelationExists   = New(CodeRelationExists, "关系已存在")
	ErrBlockedUser      = New(CodeBlockedUser, "用户已被拉黑")
	ErrInvalidFollowing = New(CodeSelfRelation, "不能关注自己").WithStatus(http.StatusBadRequest)

	// 地理位置错误
	ErrInvalidLocation  = New(CodeInvalidLocation, "无效的位置坐标").WithStatus(http.StatusBadRequest)
	ErrLocationDisabled = New(CodeLocationDisabled, "位置服务已禁用").WithStatus(http.StatusForbidden)
	ErrLocationNotFound = New(CodeLocationNotFound, "位置信息不存在").WithStatus(http.StatusNotFound)

	// 聊天相关错误
	ErrSelfChat           = New(CodeSelfChat, "不能与自己创建私聊").WithStatus(http.StatusBadRequest)
	ErrNotChatMember      = New(CodeNotChatMember, "不是聊天室成员").WithStatus(http.StatusForbidden)
	ErrChatRoomNotFound   = New(CodeChatRoomNotFound, "聊天室不存在").WithStatus(http.StatusNotFound)
	ErrMessageNotFound    = New(CodeMessageNotFound, "消息不存在").WithStatus(http.StatusNotFound)
	ErrInvalidMessageType = New(CodeMessageTypeInvalid, "无效的消息类型").WithStatus(http.StatusBadRequest)
	// 话题相关错误
	ErrTopicNotFound = New(CodeTopicNotFound, "话题不存在")
	ErrTopicExpired  = New(CodeTopicExpired, "话题已过期")
	ErrTagNotFound   = New(CodeTagNotFound, "标签不存在")
)
