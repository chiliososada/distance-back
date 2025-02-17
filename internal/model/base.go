package model

import (
	"mime/multipart"
	"time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	UID       string    `gorm:"column:uid;type:varchar(36);primaryKey;not null" json:"uid"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at" json:"updated_at"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.UID == "" {
		base.UID = uuid.NewV4().String()
	}
	return nil
}

// Pagination 分页参数
type Pagination struct {
	Page     int   `json:"page" form:"page"`
	PageSize int   `json:"page_size" form:"page_size"`
	Total    int64 `json:"total"`
}

// Location 位置信息
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// File 文件信息结构体
type File struct {
	File   *multipart.FileHeader `json:"-"`      // 文件对象
	Type   string                `json:"type"`   // 文件类型
	Name   string                `json:"name"`   // 文件名
	Size   uint                  `json:"size"`   // 文件大小
	Width  int                   `json:"width"`  // 图片宽度（仅图片类型）
	Height int                   `json:"height"` // 图片高度（仅图片类型）
}
