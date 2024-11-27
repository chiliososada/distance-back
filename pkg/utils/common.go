package utils

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"mime/multipart"
	"path"
	"regexp"
	"strings"
	"time"
)

// StringToJSON 将字符串解析为JSON对象
func StringToJSON(str string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// NewNullString 创建 sql.NullString
func NewNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// JSONToString 将JSON对象转换为字符串
func JSONToString(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ValidateFile 验证文件
func ValidateFile(file *multipart.FileHeader, maxSize int64, allowedTypes string) error {
	// 检查文件大小
	if file.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	// 检查文件类型
	ext := strings.ToLower(path.Ext(file.Filename))
	if !strings.Contains(allowedTypes, ext) {
		return fmt.Errorf("file type %s is not allowed", ext)
	}

	return nil
}

// MD5 计算字符串的MD5值
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// IsEmail 检查是否是有效的邮箱地址
func IsEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// IsPhone 检查是否是有效的手机号（国际格式）
func IsPhone(phone string) bool {
	pattern := `^\+?[1-9]\d{1,14}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(phone)
}

// // TruncateString 截断字符串到指定长度，并添加省略号
// func TruncateString(str string, length int) string {
// 	if length <= 0 {
// 		return ""
// 	}

// 	runes := []rune(str)
// 	if len(runes) <= length {
// 		return str
// 	}

// 	return string(runes[:length]) + "..."
// }

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// RemoveWhiteSpace 删除字符串中的所有空白字符
func RemoveWhiteSpace(str string) string {
	return strings.Join(strings.Fields(str), "")
}

// FormatFileSize 格式化文件大小
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatTimestamp 格式化时间戳
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseTimestamp 解析时间戳
func ParseTimestamp(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

// IsEmptyString 检查字符串是否为空
func IsEmptyString(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// ExtractFileExt 提取文件扩展名
func ExtractFileExt(filename string) string {
	return strings.ToLower(path.Ext(filename))
}

// TimePtr 将 time.Time 转换为 *time.Time
func TimePtr(t time.Time) *time.Time {
	return &t
}
