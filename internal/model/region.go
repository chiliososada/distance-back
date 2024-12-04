package model

// Region 地区模型
type Region struct {
	BaseModel
	ParentUID string  `gorm:"type:varchar(36)" json:"parent_uid"`
	Name      string  `gorm:"size:50" json:"name"`
	Level     string  `gorm:"type:enum('country','province','city','district')" json:"level"`
	Code      string  `gorm:"size:20;uniqueIndex" json:"code"`
	Latitude  float64 `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude float64 `gorm:"type:decimal(11,8)" json:"longitude"`
	Parent    *Region `gorm:"foreignKey:ParentUID;references:UID" json:"parent"`
}

// UserLocation 用户位置历史模型
type UserLocation struct {
	BaseModel
	UserUID   string  `gorm:"type:varchar(36);index" json:"user_uid"`
	Latitude  float64 `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude float64 `gorm:"type:decimal(11,8)" json:"longitude"`
	RegionUID string  `gorm:"type:varchar(36)" json:"region_uid"`
	User      User    `gorm:"foreignKey:UserUID;references:UID" json:"user"`
	Region    Region  `gorm:"foreignKey:RegionUID;references:UID" json:"region"`
}
