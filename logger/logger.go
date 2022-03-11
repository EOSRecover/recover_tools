package logger

import (
	"fmt"
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"strings"
	"time"
)

var logger *zap.SugaredLogger

func InitLogger(name string) (err error) {

	filename := fmt.Sprintf("./logs/%s.log", name)

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "linenum",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.DebugLevel)

	writer, err := getWriter(filename)
	if err != nil {

		return
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(writer), zapcore.AddSync(os.Stdout)),
		atomicLevel,
	)

	caller := zap.AddCaller()
	development := zap.Development()

	lg := zap.New(core, caller, development)
	logger = lg.Sugar()
	return
}

func Instance() *zap.SugaredLogger {

	return logger
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("[2006-01-02 15:04:05]"))
}

func getWriter(filename string) (hook io.Writer, err error) {

	name := strings.Replace(filename, ".log", "", -1) + "_%Y-%m-%d.log"
	hook, err = rotateLogs.New(
		name,
		rotateLogs.WithRotationCount(8), // 文件最大保存份数
		rotateLogs.WithRotationTime(24*time.Hour), //文件切割时间
	)

	if err != nil {

		fmt.Println(err)
		return
	}

	return
}
