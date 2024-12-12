package log

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	xlogger "gorm.io/gorm/logger"
)

const (
	// Silent silent log level
	Silent int = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
)

type GormLogger struct {
	logger   LogCore
	logLevel xlogger.LogLevel
}

func NewGormLogger(l LogCore, level int) xlogger.Interface {
	return &GormLogger{
		logger:   l,
		logLevel: xlogger.LogLevel(level),
	}
}

func (g *GormLogger) LogMode(LogLevel xlogger.LogLevel) xlogger.Interface {
	newLogger := *g
	newLogger.logLevel = LogLevel
	return &newLogger
}

func (g GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if g.logLevel >= xlogger.Info {
		g.logger.InfoWithCtx(ctx, fmt.Sprintf("%s %s", trimPath(), fmt.Sprint(data...)))
	}
}

func (g GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if g.logLevel >= xlogger.Warn {
		g.logger.WarnWithCtx(ctx, fmt.Sprintf("%s %s", trimPath(), fmt.Sprint(data...)))
	}
}

func (g GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if g.logLevel >= xlogger.Error {
		g.logger.ErrorWithCtx(ctx, fmt.Sprintf("%s %s", trimPath(), fmt.Sprint(data...)))
	}
}

func (g GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil:
		g.logger.ErrorWithCtx(ctx, trimPath(), Any("err", err.Error()), Any("elapsed", fmt.Sprintf("%vms", float64(elapsed.Nanoseconds())/1e6)), Any("rows", rows), Any("sql", sql))
	default:
		g.logger.InfoWithCtx(ctx, trimPath(), Any("elapsed", fmt.Sprintf("%vms", float64(elapsed.Nanoseconds())/1e6)), Any("rows", rows), Any("sql", sql))
	}
}

func trimPath() string {
	f := ""
	for i := 4; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok {
			f = file + ":" + strconv.FormatInt(int64(line), 10)
			break
		}
	}
	if f == "" {
		return f
	}
	idx := strings.LastIndex(f, "/")
	if idx == -1 {
		return f
	}
	idx = strings.LastIndex(f[:idx], "/")
	if idx == -1 {
		return f
	}
	return f[idx+1:]
}
