package request

import "time"

// PaginationQuery 分页查询参数
type PaginationQuery struct {
	Page int `form:"page" binding:"required,min=1" json:"page"`         // 页码
	Size int `form:"size" binding:"required,min=1,max=100" json:"size"` // 每页大小
}

// SortQuery 排序查询参数
type SortQuery struct {
	SortBy    string `form:"sort_by" json:"sort_by"`       // 排序字段
	SortOrder string `form:"sort_order" json:"sort_order"` // 排序方向(asc/desc)
}

// TimeRangeQuery 时间范围查询参数
type TimeRangeQuery struct {
	StartTime time.Time `form:"start_time" binding:"omitempty" time_format:"2006-01-02 15:04:05" json:"start_time"`               // 开始时间
	EndTime   time.Time `form:"end_time" binding:"omitempty,gtfield=StartTime" time_format:"2006-01-02 15:04:05" json:"end_time"` // 结束时间
}

// DateRangeQuery 日期范围查询参数
type DateRangeQuery struct {
	StartDate string `form:"start_date" binding:"omitempty,datetime=2006-01-02" json:"start_date"`                // 开始日期
	EndDate   string `form:"end_date" binding:"omitempty,datetime=2006-01-02,gtefield=StartDate" json:"end_date"` // 结束日期
}

// LocationQuery 位置查询参数
type LocationQuery struct {
	Latitude  float64 `form:"latitude" binding:"required,min=-90,max=90" json:"latitude"`     // 纬度
	Longitude float64 `form:"longitude" binding:"required,min=-180,max=180" json:"longitude"` // 经度
	Radius    float64 `form:"radius" binding:"required,min=0,max=50000" json:"radius"`        // 搜索半径(米)
}
type LocationPerson struct {
	Latitude  float64 `form:"latitude" binding:"required,min=-90,max=90" json:"latitude"`     // 纬度
	Longitude float64 `form:"longitude" binding:"required,min=-180,max=180" json:"longitude"` // 经度

}

// ValidateSortOrder 验证排序方向
func (q *SortQuery) ValidateSortOrder() string {
	if q.SortOrder != "asc" && q.SortOrder != "desc" {
		return "desc"
	}
	return q.SortOrder
}

// GetTimeRange 获取时间范围，如果未指定则返回默认值
func (q *TimeRangeQuery) GetTimeRange() (time.Time, time.Time) {
	if q.StartTime.IsZero() {
		q.StartTime = time.Now().AddDate(0, -1, 0) // 默认近一个月
	}
	if q.EndTime.IsZero() {
		q.EndTime = time.Now()
	}
	return q.StartTime, q.EndTime
}

// GetDateRange 获取日期范围，如果未指定则返回默认值
func (q *DateRangeQuery) GetDateRange() (time.Time, time.Time, error) {
	var startDate, endDate time.Time
	var err error

	if q.StartDate == "" {
		startDate = time.Now().AddDate(0, -1, 0) // 默认近一个月
	} else {
		startDate, err = time.Parse("2006-01-02", q.StartDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if q.EndDate == "" {
		endDate = time.Now()
	} else {
		endDate, err = time.Parse("2006-01-02", q.EndDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	// 设置日期范围的开始和结束时间
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.Local)

	return startDate, endDate, nil
}

// UIDParam UUID参数
type UIDParam struct {
	UID string `uri:"uid" binding:"required,uuid" json:"uid"` // UUID参数
}
