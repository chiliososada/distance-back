package handler

import (
	"net/http"
	"strconv"

	"DistanceBack_v1/internal/service"
	"DistanceBack_v1/pkg/errors" // 添加这个导入

	"github.com/gin-gonic/gin"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationQuery 分页查询参数
type PaginationQuery struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100"`
}

// Handler 处理器基础结构
type Handler struct {
	userService         *service.UserService
	topicService        *service.TopicService
	chatService         *service.ChatService
	relationshipService *service.RelationshipService
}

// NewHandler 创建处理器实例
func NewHandler(
	userService *service.UserService,
	topicService *service.TopicService,
	chatService *service.ChatService,
	relationshipService *service.RelationshipService,
) *Handler {
	return &Handler{
		userService:         userService,
		topicService:        topicService,
		chatService:         chatService,
		relationshipService: relationshipService,
	}
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, err error) {
	// 处理应用错误
	if e, ok := err.(*errors.AppError); ok { // 修改这里
		c.JSON(e.HTTPStatus, Response{
			Code:    e.Code,
			Message: e.Message,
			Data:    e.Details,
		})
		return
	}

	// 处理其他错误
	c.JSON(http.StatusInternalServerError, Response{
		Code:    500,
		Message: err.Error(),
	})
}

// GetCurrentUserID 获取当前用户ID
func (h *Handler) GetCurrentUserID(c *gin.Context) uint64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint64)
}

// ParseUint64Param 解析uint64类型的路径参数
func ParseUint64Param(c *gin.Context, param string) (uint64, error) {
	val := c.Param(param)
	return strconv.ParseUint(val, 10, 64)
}

// GetPagination 获取分页参数
func GetPagination(c *gin.Context) (*PaginationQuery, error) {
	var query PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		return nil, err
	}
	return &query, nil
}
