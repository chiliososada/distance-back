package auth

import (
	"firebase.google.com/go/v4/auth"
)

// AuthUser 认证用户模型
type AuthUser struct {
	UID         string                 `json:"uid"`
	Email       string                 `json:"email,omitempty"`
	PhoneNumber string                 `json:"phone_number,omitempty"`
	DisplayName string                 `json:"display_name,omitempty"`
	PhotoURL    string                 `json:"photo_url,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Claims      map[string]interface{} `json:"claims,omitempty"`
}

// NewAuthUser 从Firebase用户记录创建认证用户
func NewAuthUser(user *auth.UserRecord) *AuthUser {
	return &AuthUser{
		UID:         user.UID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		DisplayName: user.DisplayName,
		PhotoURL:    user.PhotoURL,
		Claims:      user.CustomClaims,
	}
}

// HasRole 检查用户是否具有指定角色
func (u *AuthUser) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsAdmin 检查是否是管理员
func (u *AuthUser) IsAdmin() bool {
	return u.HasRole("admin")
}

// IsMerchant 检查是否是商家
func (u *AuthUser) IsMerchant() bool {
	return u.HasRole("merchant")
}

// GetUID 获取用户UID
func (u *AuthUser) GetUID() string {
	return u.UID
}

// GetEmail 获取用户邮箱
func (u *AuthUser) GetEmail() string {
	return u.Email
}

// GetPhoneNumber 获取用户手机号
func (u *AuthUser) GetPhoneNumber() string {
	return u.PhoneNumber
}
