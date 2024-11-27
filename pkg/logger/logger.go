package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

type Options struct {
	LogFileDir    string
	AppName       string
	ErrorFileName string
	WarnFileName  string
	InfoFileName  string
	DebugFileName string
	MaxSize       int  // megabytes
	MaxBackups    int  // number of backups
	MaxAge        int  // days
	Development   bool // 是否开发模式
}

func InitLogger(opt *Options) {
	// 设置默认参数
	if opt.LogFileDir == "" {
		opt.LogFileDir = "./logs"
	}
	if opt.AppName == "" {
		opt.AppName = "app"
	}
	if opt.MaxSize == 0 {
		opt.MaxSize = 100
	}
	if opt.MaxBackups == 0 {
		opt.MaxBackups = 60
	}
	if opt.MaxAge == 0 {
		opt.MaxAge = 30
	}

	// 创建日志目录
	if err := os.MkdirAll(opt.LogFileDir, os.ModePerm); err != nil {
		panic(fmt.Errorf("创建日志目录失败: %v", err))
	}

	// 创建核心记录器
	cores := make([]zapcore.Core, 0, 4)
	encoder := getEncoder()
	level := zap.DebugLevel
	if !opt.Development {
		level = zap.InfoLevel
	}

	// error 级别
	errorWriter := getLogWriter(fmt.Sprintf("%s/%s_error.log", opt.LogFileDir, opt.AppName),
		opt.MaxSize, opt.MaxBackups, opt.MaxAge)
	cores = append(cores, zapcore.NewCore(encoder, errorWriter, zap.ErrorLevel))

	// warn 级别
	warnWriter := getLogWriter(fmt.Sprintf("%s/%s_warn.log", opt.LogFileDir, opt.AppName),
		opt.MaxSize, opt.MaxBackups, opt.MaxAge)
	cores = append(cores, zapcore.NewCore(encoder, warnWriter, zap.WarnLevel))

	// info 级别
	infoWriter := getLogWriter(fmt.Sprintf("%s/%s_info.log", opt.LogFileDir, opt.AppName),
		opt.MaxSize, opt.MaxBackups, opt.MaxAge)
	cores = append(cores, zapcore.NewCore(encoder, infoWriter, zap.InfoLevel))

	// debug 级别
	debugWriter := getLogWriter(fmt.Sprintf("%s/%s_debug.log", opt.LogFileDir, opt.AppName),
		opt.MaxSize, opt.MaxBackups, opt.MaxAge)
	cores = append(cores, zapcore.NewCore(encoder, debugWriter, level))

	// 开发模式下同时输出到控制台
	if opt.Development {
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level))
	}

	core := zapcore.NewTee(cores...)
	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// 以下是包装的日志方法
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}
