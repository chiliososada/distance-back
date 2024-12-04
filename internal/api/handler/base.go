package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/internal/service"
	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/chiliososada/distance-back/pkg/logger"

	"github.com/gin-gonic/gin"
)

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
	c.JSON(http.StatusOK, response.SuccessResponse(data))
}

// Error 返回错误响应
func Error(c *gin.Context, err error) {
	// 处理应用错误
	if e, ok := err.(*errors.AppError); ok {
		logger.Error("Application error",
			logger.String("path", c.Request.URL.Path),
			logger.Int("code", e.Code),
			logger.String("message", e.Message),
			logger.Any("details", e.Details),
			logger.String("developer", e.Developer))

		c.JSON(e.HTTPStatus, response.Response{
			Code:    e.Code,
			Message: e.Message,
			Data:    e.Details,
			TraceID: GetTraceID(c),
		})
		return
	}

	// 处理其他错误
	logger.Error("Internal server error",
		logger.String("path", c.Request.URL.Path),
		logger.Any("error", err))

	c.JSON(http.StatusInternalServerError, response.Response{
		Code:    errors.CodeUnknown,
		Message: "Internal server error",
		TraceID: GetTraceID(c),
	})
}

// GetCurrentUserUID 获取当前用户UID
func (h *Handler) GetCurrentUserUID(c *gin.Context) string {
	firebaseUID, exists := c.Get("user_uid")
	if !exists {
		return ""
	}

	if firebaseUIDStr, ok := firebaseUID.(string); ok {
		// 通过 Firebase UID 查询用户
		user, err := h.userService.GetByFirebaseUID(c.Request.Context(), firebaseUIDStr)
		if err != nil {
			logger.Error("Failed to get user by firebase uid",
				logger.String("firebase_uid", firebaseUIDStr),
				logger.Any("error", err))
			return ""
		}
		if user == nil {
			return ""
		}
		return user.UID
	}

	return ""
}

// ParseUUID 解析UUID类型的路径参数
func ParseUUID(c *gin.Context, param string) (string, error) {
	uid := c.Param(param)
	if uid == "" {
		return "", errors.New(errors.CodeValidation, "Invalid UUID parameter").
			WithDetails(map[string]string{
				"param": param,
			})
	}
	return uid, nil
}

// ParseDateRange 解析日期范围
func ParseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New(errors.CodeValidation, "Invalid start date format").
				WithDetails(map[string]string{
					"start_date": startDate,
					"format":     "2006-01-02",
				})
		}
	} else {
		start = time.Now().AddDate(0, -1, 0)
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New(errors.CodeValidation, "Invalid end date format").
				WithDetails(map[string]string{
					"end_date": endDate,
					"format":   "2006-01-02",
				})
		}
	} else {
		end = time.Now()
	}

	// 验证日期范围
	if end.Before(start) {
		return time.Time{}, time.Time{}, errors.New(errors.CodeValidation, "End date must be after start date").
			WithDetails(map[string]string{
				"start_date": start.Format("2006-01-02"),
				"end_date":   end.Format("2006-01-02"),
			})
	}

	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, time.Local)

	return start, end, nil
}

// GetPagination 获取分页参数
func GetPagination(c *gin.Context) (page, size int, err error) {
	pageStr := c.DefaultQuery("page", "1")
	page, err = strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, 0, errors.New(errors.CodeValidation, "Invalid page parameter").
			WithDetails(map[string]interface{}{
				"page": pageStr,
				"min":  1,
			})
	}

	sizeStr := c.DefaultQuery("size", "10")
	size, err = strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		return 0, 0, errors.New(errors.CodeValidation, "Invalid size parameter").
			WithDetails(map[string]interface{}{
				"size":        sizeStr,
				"valid_range": "1 to 100",
			})
	}

	return page, size, nil
}

// GetSort 获取排序参数
func GetSort(c *gin.Context) (string, string) {
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// 验证排序方向
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return sortBy, sortOrder
}

// ValidateLocation 验证位置信息
func ValidateLocation(latitude, longitude, radius float64) error {
	if latitude < -90 || latitude > 90 {
		return errors.New(errors.CodeInvalidLocation, "Invalid latitude").
			WithDetails(map[string]interface{}{
				"latitude":    latitude,
				"valid_range": "-90 to 90",
			})
	}
	if longitude < -180 || longitude > 180 {
		return errors.New(errors.CodeInvalidLocation, "Invalid longitude").
			WithDetails(map[string]interface{}{
				"longitude":   longitude,
				"valid_range": "-180 to 180",
			})
	}
	if radius < 0 || radius > 50000 {
		return errors.New(errors.CodeInvalidLocation, "Invalid radius").
			WithDetails(map[string]interface{}{
				"radius":      radius,
				"valid_range": "0 to 50000",
			})
	}
	return nil
}

// GetTraceID 获取请求追踪ID
func GetTraceID(c *gin.Context) string {
	return c.GetHeader("X-Request-ID")
}

// BindAndValidate 绑定并验证请求参数
func BindAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBind(obj); err != nil {
		return errors.New(errors.CodeValidation, "Invalid request parameters").
			WithDetails(map[string]interface{}{
				"error": err.Error(),
			})
	}
	return nil
}

// BindQuery 绑定并验证查询参数
func BindQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.New(errors.CodeValidation, "Invalid query parameters").
			WithDetails(map[string]interface{}{
				"error": err.Error(),
			})
	}
	return nil
}
