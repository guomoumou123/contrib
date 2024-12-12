package log

import (
	"context"
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Field struct {
	key   string
	value interface{}
}

type LogCore interface {
	SetPrefix(msg string) LogCore
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
	Fatal(args ...interface{})
	InfoWithCtx(ctx context.Context, msg string, fields ...Field)
	ErrorWithCtx(ctx context.Context, msg string, fields ...Field)
	WarnWithCtx(ctx context.Context, msg string, fields ...Field)
	FatalWithCtx(ctx context.Context, msg string, fields ...Field)
}

type Config struct {
	Debug       bool   //是否开启调试
	MaxSize     int    //日志文件最大多少兆
	MaxAge      int    //日志文件保留天数
	MaxBackups  int    //保留文件数
	FileName    string //日志名字
	Compress    bool   //日志生成压缩包,大幅降低磁盘空间,必要时使用
	RotateByDay bool   //每天轮转一次,如果开启,maxBackups的值需要>=maxDays
}

func initWriter(conf *Config) io.Writer {
	if conf == nil {
		conf = defaultConfig()
	}

	if conf.FileName == "" {
		conf.FileName = defaultFileName
	}

	if conf.MaxSize == 0 {
		conf.MaxSize = defaultMaxSize
	}

	if conf.MaxAge == 0 {
		conf.MaxAge = defaultMaxAge
	}

	if conf.MaxBackups == 0 {
		conf.MaxBackups = defaultMaxBackups
	}

	return &lumberjack.Logger{
		Filename:   conf.FileName,
		MaxSize:    conf.MaxSize,
		MaxAge:     conf.MaxAge,
		MaxBackups: conf.MaxBackups,
		LocalTime:  true,
		Compress:   conf.Compress,
	}
}

func NewLogger(lc *Config, loggerType string) LogCore {
	switch loggerType {
	case "file":
		w := initWriter(lc)
		return newZapLogger(w, lc.Debug)
	case "stdout":
		return newZapLogger(os.Stdout, true)
	}
	return nil
}

func Any(key string, value interface{}) Field {
	return Field{
		key:   key,
		value: value,
	}
}
