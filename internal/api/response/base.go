// internal/api/response/base.go
package response

// Response 基础响应结构
type Response struct {
	Code    int         `json:"code"`               // 状态码
	Message string      `json:"message"`            // 响应消息
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	List    interface{} `json:"list"`     // 数据列表
	Total   int64       `json:"total"`    // 总记录数
	Page    int         `json:"page"`     // 当前页码
	Size    int         `json:"size"`     // 每页大小
	Pages   int         `json:"pages"`    // 总页数
	HasMore bool        `json:"has_more"` // 是否有更多数据
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      int         `json:"code"`                 // 错误码
	Message   string      `json:"message"`              // 错误消息
	Details   interface{} `json:"details,omitempty"`    // 错误详情
	RequestID string      `json:"request_id,omitempty"` // 请求ID
}

// ValidationError 验证错误结构
type ValidationError struct {
	Field   string `json:"field"`   // 字段名
	Message string `json:"message"` // 错误信息
}

// New 创建标准响应
func New(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewError 创建错误响应
func NewError(code int, message string, details interface{}) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewPaginated 创建分页响应
func NewPaginated(list interface{}, total int64, page, size int) *PaginatedResponse {
	pages := int(total) / size
	if int(total)%size > 0 {
		pages++
	}

	return &PaginatedResponse{
		List:    list,
		Total:   total,
		Page:    page,
		Size:    size,
		Pages:   pages,
		HasMore: page < pages,
	}
}
