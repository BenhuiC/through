package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"through/config"
)

var defLogger *Logger

var logConfig zap.Config

func Init() (err error) {
	cfg := config.Common

	if cfg.Env == "prod" {
		logConfig = zap.NewProductionConfig()
	} else {
		logConfig = zap.NewDevelopmentConfig()
	}
	if cfg.LogFile != "" {
		logConfig.OutputPaths = append(logConfig.OutputPaths, cfg.LogFile)
		logConfig.ErrorOutputPaths = append(logConfig.ErrorOutputPaths, cfg.LogFile)
	}

	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	defLogger = NewLogger(zap.AddCallerSkip(1))
	return
}

func Infof(format string, args ...interface{}) {
	defLogger.Infof(format, args...)
}

func Info(args ...interface{}) {
	defLogger.Info(args...)
}

func Warnf(format string, args ...interface{}) {
	defLogger.Warnf(format, args...)
}

func Warn(args ...interface{}) {
	defLogger.Warn(args...)
}

func Errorf(format string, args ...interface{}) {
	defLogger.Errorf(format, args...)
}

func Error(args ...interface{}) {
	defLogger.Error(args...)
}

func Debugf(format string, args ...interface{}) {
	defLogger.Debugf(format, args...)
}

func Debug(args ...interface{}) {
	defLogger.Debug(args...)
}
