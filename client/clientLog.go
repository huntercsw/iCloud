package main

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func getLogLevel() zapcore.LevelEnabler {
	switch Conf.Level {
	case "error":
		return zapcore.ErrorLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	default:
		return zapcore.DebugLevel
	}
}

func InitLogger() {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, getLogLevel())

	logger := zap.New(core, zap.AddCaller())
	Logger = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)

	// log format to json
	// return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   Conf.Log,
		MaxSize:    Conf.MaxSize,
		MaxBackups: Conf.MaxBackups,		// 保留旧文件的最大个数
		MaxAge:     Conf.MaxAge,		// 保留旧文件的最大天数
		Compress:   Conf.Compress,	// 是否压缩/归档旧文件
	}
	return zapcore.AddSync(lumberJackLogger)
}