package response

// Response 基础响应结构
type Response struct {
	Code    int         `json:"code"`               // 状态码
	Message string      `json:"message"`            // 响应消息
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// Location 位置信息
type Location struct {
	Latitude  float64 `json:"latitude"`           // 纬度
	Longitude float64 `json:"longitude"`          // 经度
	Distance  float64 `json:"distance,omitempty"` // 距离(米)
}

// Pagination 分页信息
type Pagination struct {
	Total   int64 `json:"total"`    // 总记录数
	Page    int   `json:"page"`     // 当前页码
	Size    int   `json:"size"`     // 每页大小
	HasMore bool  `json:"has_more"` // 是否有更多数据
}

// PaginatedResponse 分页响应结构
type PaginatedResponse[T any] struct {
	List       []T        `json:"list"`       // 数据列表
	Pagination Pagination `json:"pagination"` // 分页信息
}

// NewResponse 创建标准响应
func NewResponse(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// SuccessResponse 创建成功响应
func SuccessResponse(data interface{}) *Response {
	return NewResponse(0, "success", data)
}

// ErrorResponse 创建错误响应
func ErrorResponse(code int, message string) *Response {
	return NewResponse(code, message, nil)
}

// NewPaginatedResponse 创建分页响应
func NewPaginatedResponse[T any](list []T, total int64, page, size int) *PaginatedResponse[T] {
	hasMore := int64((page)*size) < total

	return &PaginatedResponse[T]{
		List: list,
		Pagination: Pagination{
			Total:   total,
			Page:    page,
			Size:    size,
			HasMore: hasMore,
		},
	}
}

// EmptyPaginatedResponse 创建空的分页响应
func EmptyPaginatedResponse[T any]() *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		List: make([]T, 0),
		Pagination: Pagination{
			Total:   0,
			Page:    1,
			Size:    10,
			HasMore: false,
		},
	}
}
