package model

import "time"

// NotificationTemplate 通知模板模型
type NotificationTemplate struct {
	BaseModel
	Code            string `gorm:"size:50;uniqueIndex" json:"code"`
	TitleTemplate   string `gorm:"type:text" json:"title_template"`
	ContentTemplate string `gorm:"type:text" json:"content_template"`
	Platform        string `gorm:"type:enum('all','ios','android','web');default:'all'" json:"platform"`
	Variables       JSON   `gorm:"type:json" json:"variables"`
	Status          string `gorm:"type:enum('active','inactive');default:'active'" json:"status"`
}

// PushNotification 推送消息模型
type PushNotification struct {
	BaseModel
	TemplateUID  string               `gorm:"type:varchar(36)" json:"template_uid"`
	Title        string               `gorm:"size:255" json:"title"`
	Content      string               `gorm:"type:text" json:"content"`
	Type         string               `gorm:"type:enum('system','merchant','topic','chat')" json:"type"`
	SenderUID    string               `gorm:"type:varchar(36)" json:"sender_uid"`
	TargetType   string               `gorm:"type:enum('all','area','specific_users')" json:"target_type"`
	TargetParams JSON                 `gorm:"type:json" json:"target_params"`
	Status       string               `gorm:"type:enum('draft','scheduled','sending','sent','cancelled')" json:"status"`
	ScheduledAt  time.Time            `json:"scheduled_at"`
	SentAt       time.Time            `json:"sent_at"`
	Template     NotificationTemplate `gorm:"foreignKey:TemplateUID;references:UID" json:"template"`
	Sender       User                 `gorm:"foreignKey:SenderUID;references:UID" json:"sender"`
}

// PushNotificationRecipient 推送接收记录模型
type PushNotificationRecipient struct {
	BaseModel
	NotificationUID string           `gorm:"type:varchar(36);uniqueIndex:uk_notification_user" json:"notification_uid"`
	UserUID         string           `gorm:"type:varchar(36);uniqueIndex:uk_notification_user" json:"user_uid"`
	DeviceUID       string           `gorm:"type:varchar(36)" json:"device_uid"`
	Status          string           `gorm:"type:enum('pending','sent','failed','received','read')" json:"status"`
	ErrorMessage    string           `gorm:"type:text" json:"error_message"`
	SentAt          time.Time        `json:"sent_at"`
	ReceivedAt      time.Time        `json:"received_at"`
	ReadAt          time.Time        `json:"read_at"`
	Notification    PushNotification `gorm:"foreignKey:NotificationUID;references:UID" json:"notification"`
	User            User             `gorm:"foreignKey:UserUID;references:UID" json:"user"`
	Device          UserDevice       `gorm:"foreignKey:DeviceUID;references:UID" json:"device"`
}

// JSON 自定义JSON类型
type JSON map[string]interface{}
