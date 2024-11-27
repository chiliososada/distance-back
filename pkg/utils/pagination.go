package utils

import (
	"math"
)

// Pagination 分页结构
type Pagination struct {
	CurrentPage int   `json:"current_page"` // 当前页码
	PerPage     int   `json:"per_page"`     // 每页条数
	Total       int64 `json:"total"`        // 总记录数
	LastPage    int   `json:"last_page"`    // 最后一页页码
	From        int   `json:"from"`         // 当前页第一条记录的序号
	To          int   `json:"to"`           // 当前页最后一条记录的序号
	HasMore     bool  `json:"has_more"`     // 是否还有更多数据
}

const (
	DefaultPage    = 1
	DefaultPerPage = 10
	MaxPerPage     = 100
)

// NewPagination 创建分页实例
func NewPagination(page, perPage int, total int64) *Pagination {
	// 验证页码
	if page < 1 {
		page = DefaultPage
	}

	// 验证每页条数
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	// 计算最后一页页码
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	// 确保当前页不超过最后一页
	if page > lastPage {
		page = lastPage
	}

	// 计算当前页数据的范围
	from := (page-1)*perPage + 1
	to := page * perPage
	if int64(to) > total {
		to = int(total)
	}

	return &Pagination{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
		From:        from,
		To:          to,
		HasMore:     page < lastPage,
	}
}

// GetOffset 获取数据库查询的偏移量
func (p *Pagination) GetOffset() int {
	return (p.CurrentPage - 1) * p.PerPage
}

// GetLimit 获取数据库查询的限制数
func (p *Pagination) GetLimit() int {
	return p.PerPage
}

// GetPageInfo 获取分页信息
func (p *Pagination) GetPageInfo() map[string]interface{} {
	return map[string]interface{}{
		"current_page": p.CurrentPage,
		"per_page":     p.PerPage,
		"total":        p.Total,
		"last_page":    p.LastPage,
		"from":         p.From,
		"to":           p.To,
		"has_more":     p.HasMore,
	}
}
