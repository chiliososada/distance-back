package response

// Response 基础响应结构
type Response struct {
	Code    int         `json:"code"`               // 状态码
	Message string      `json:"message"`            // 响应消息
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// PaginatedResponse 分页响应结构
type PaginatedResponse[T any] struct {
	List    []T   `json:"list"`     // 数据列表
	Total   int64 `json:"total"`    // 总记录数
	Page    int   `json:"page"`     // 当前页码
	Size    int   `json:"size"`     // 每页大小
	HasMore bool  `json:"has_more"` // 是否还有更多数据
}

// NewResponse 创建标准响应
func NewResponse(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewPaginatedResponse 创建分页响应
func NewPaginatedResponse[T any](list []T, total int64, page, size int) *PaginatedResponse[T] {
	hasMore := int64((page)*size) < total

	return &PaginatedResponse[T]{
		List:    list,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: hasMore,
	}
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *Response {
	return NewResponse(0, "success", data)
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) *Response {
	return NewResponse(code, message, nil)
}
