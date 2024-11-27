package logger

import (
	"time"

	"go.uber.org/zap"
)

// 字符串
func String(key string, value string) zap.Field {
	return zap.String(key, value)
}

// 整型
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64
func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Uint64
func Uint64(key string, value uint64) zap.Field {
	return zap.Uint64(key, value)
}

// Float64
func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

// Bool
func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// Error
func Err(err error) zap.Field {
	return zap.Error(err)
}

// Any 用于记录任意类型
func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// Duration 用于记录时间间隔
func Duration(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

// Time 用于记录时间点
func Time(key string, value time.Time) zap.Field {
	return zap.Time(key, value)
}

// Strings 用于记录字符串切片
func Strings(key string, value []string) zap.Field {
	return zap.Strings(key, value)
}

// ByteString 用于记录字节切片为字符串
func ByteString(key string, value []byte) zap.Field {
	return zap.ByteString(key, value)
}

// Bools 用于记录布尔切片
func Bools(key string, value []bool) zap.Field {
	return zap.Bools(key, value)
}

// Stack 用于记录堆栈信息
func Stack() zap.Field {
	return zap.Stack("stack")
}

// Binary 用于记录二进制数据
func Binary(key string, value []byte) zap.Field {
	return zap.Binary(key, value)
}

// NamedError 用于记录带名称的错误
func NamedError(key string, err error) zap.Field {
	return zap.NamedError(key, err)
}
