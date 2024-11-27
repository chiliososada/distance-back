package request

import "time"

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page" form:"page" binding:"required,min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"required,min=1,max=100"`
}

// TimeRange 时间范围
type TimeRange struct {
	StartTime time.Time `json:"start_time" form:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" form:"end_time" binding:"required,gtfield=StartTime"`
}

// Sort 排序参数
type Sort struct {
	Field     string `json:"field" form:"field" binding:"required"`
	Direction string `json:"direction" form:"direction" binding:"required,oneof=asc desc"`
}

// Location 位置参数
type Location struct {
	Latitude  float64 `json:"latitude" form:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" form:"longitude" binding:"required,min=-180,max=180"`
	Radius    float64 `json:"radius" form:"radius" binding:"required,min=0,max=50000"` // 米为单位，最大50km
}

// DateRange 日期范围
type DateRange struct {
	StartDate string `json:"start_date" form:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate   string `json:"end_date" form:"end_date" binding:"required,datetime=2006-01-02,gtfield=StartDate"`
}
