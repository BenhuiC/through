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

func Info(format string, args ...interface{}) {
	defLogger.Infof(format, args...)
}

func Warn(format string, args ...interface{}) {
	defLogger.Warnf(format, args...)
}

func Error(format string, args ...interface{}) {
	defLogger.Errorf(format, args)
}

func Debug(format string, args ...interface{}) {
	defLogger.Debugf(format, args)
}
