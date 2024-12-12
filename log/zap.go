package log

import (
	"context"
	"fmt"
	"io"
	"sync"

	ictx "github.com/guomoumou123/contrib/context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/utils"
)

var bufferPool = sync.Pool{New: func() any { return new([]byte) }}

type zapLogger struct {
	prefix string
	writer *zap.Logger
}

func newZapLogger(w io.Writer, debug bool) LogCore {
	ws := zapcore.AddSync(w)
	//初始化encoder配置
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeCaller = shortCallerEncoder //日志调用方法
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg), ws, zap.DebugLevel,
	)
	lz := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel), zap.AddCallerSkip(1))
	return &zapLogger{
		writer: lz,
	}
}

func shortCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + caller.TrimmedPath() + "]")
}

func (l *zapLogger) InfoWithCtx(ctx context.Context, msg string, fields ...Field) {
	f := make([]zapcore.Field, 0, len(fields)+1)
	for _, v := range fields {
		f = append(f, zap.Any(v.key, v.value))
	}
	traceId := ictx.GetTraceId(ctx)
	f = append(f, zap.String("trace_id", traceId))
	l.writer.Info(msg, f...)
}

func (l *zapLogger) ErrorWithCtx(ctx context.Context, msg string, fields ...Field) {
	f := make([]zapcore.Field, 0, len(fields)+1)
	for _, v := range fields {
		f = append(f, zap.Any(v.key, v.value))
	}
	traceId := ictx.GetTraceId(ctx)
	f = append(f, zap.String("trace_id", traceId))
	l.writer.Error(msg, f...)
}

func (l *zapLogger) WarnWithCtx(ctx context.Context, msg string, fields ...Field) {
	f := make([]zapcore.Field, 0, len(fields)+1)
	for _, v := range fields {
		f = append(f, zap.Any(v.key, v.value))
	}
	traceId := ictx.GetTraceId(ctx)
	f = append(f, zap.String("trace_id", traceId))
	l.writer.Warn(msg, f...)
}

func (l *zapLogger) FatalWithCtx(ctx context.Context, msg string, fields ...Field) {
	f := make([]zapcore.Field, 0, len(fields)+1)
	for _, v := range fields {
		f = append(f, zap.Any(v.key, v.value))
	}
	traceId := ictx.GetTraceId(ctx)
	f = append(f, zap.String("trace_id", traceId))
	l.writer.Panic(msg, f...)
}

func getBuffer() *[]byte {
	p := bufferPool.Get().(*[]byte)
	*p = (*p)[:0]
	return p
}

func putBuffer(p *[]byte) {
	if cap(*p) > 64<<10 {
		*p = nil
	}
	bufferPool.Put(p)
}

func (l *zapLogger) SetPrefix(name string) LogCore {
	newLogger := *l
	newLogger.prefix = name
	return &newLogger
}

func (l *zapLogger) Info(args ...interface{}) {
	buf := getBuffer()
	defer putBuffer(buf)
	msg := fmt.Append(*buf, args...)
	l.writer.Info(fmt.Sprintf("%s %v", l.prefix, string(msg)))
}

func (l *zapLogger) Error(args ...interface{}) {
	buf := getBuffer()
	defer putBuffer(buf)
	msg := fmt.Append(*buf, args...)
	l.writer.Error(fmt.Sprintf("%s %v", l.prefix, string(msg)))
}

func (l *zapLogger) Warn(args ...interface{}) {
	buf := getBuffer()
	defer putBuffer(buf)
	msg := fmt.Append(*buf, args...)
	l.writer.Warn(fmt.Sprintf("%s %s %v", utils.FileWithLineNum(), l.prefix, string(msg)))
}

func (l *zapLogger) Debug(args ...interface{}) {
	buf := getBuffer()
	defer putBuffer(buf)
	msg := fmt.Append(*buf, args...)
	l.writer.Debug(fmt.Sprintf("%s %v", l.prefix, string(msg)))
}

func (l *zapLogger) Fatal(args ...interface{}) {
	buf := getBuffer()
	defer putBuffer(buf)
	msg := fmt.Append(*buf, args...)
	l.writer.Fatal(fmt.Sprintf("%s %v", l.prefix, string(msg)))
}
