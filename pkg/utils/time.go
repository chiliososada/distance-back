package utils

import (
	"fmt"
	"time"
)

// 定义时间格式
const (
	TimeLayoutDate     = "2006-01-02"
	TimeLayoutTime     = "15:04:05"
	TimeLayoutDefault  = "2006-01-02 15:04:05"
	TimeLayoutCompact  = "20060102150405"
	TimeLayoutTimeZone = "2006-01-02 15:04:05 -0700"
)

// FormatTime 格式化时间为字符串
func FormatTime(t time.Time, layout string) string {
	if layout == "" {
		layout = TimeLayoutDefault
	}
	return t.Format(layout)
}

// ParseTime 解析时间字符串
func ParseTime(timeStr, layout string) (time.Time, error) {
	if layout == "" {
		layout = TimeLayoutDefault
	}
	return time.Parse(layout, timeStr)
}

// TimeAgo 计算经过时间
func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "刚刚"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d分钟前", minutes)
	case diff < time.Hour*24:
		hours := int(diff.Hours())
		return fmt.Sprintf("%d小时前", hours)
	case diff < time.Hour*24*30:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d天前", days)
	case diff < time.Hour*24*365:
		months := int(diff.Hours() / 24 / 30)
		return fmt.Sprintf("%d个月前", months)
	default:
		years := int(diff.Hours() / 24 / 365)
		return fmt.Sprintf("%d年前", years)
	}
}

// GetStartOfDay 获取某一天的开始时间
func GetStartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// GetEndOfDay 获取某一天的结束时间
func GetEndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}

// GetStartOfWeek 获取本周的开始时间（周一为第一天）
func GetStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	year, month, day := t.Date()
	return time.Date(year, month, day-weekday+1, 0, 0, 0, 0, t.Location())
}

// GetEndOfWeek 获取本周的结束时间
func GetEndOfWeek(t time.Time) time.Time {
	return GetStartOfWeek(t).AddDate(0, 0, 6)
}
