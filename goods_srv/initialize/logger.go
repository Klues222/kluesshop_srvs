package initialize

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func InitLogger() {
	file, err := os.Create("app.log")
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// 定义编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevelAt(zap.InfoLevel)

	// 构建核心（core），它决定了日志的输出位置、日志级别和编码格式
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),           // 编码器
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(file)), // 输出位置，这里设置为文件
		atomicLevel, // 日志级别
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.PanicLevel))
	zap.ReplaceGlobals(logger)
}
