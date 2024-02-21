package log

import (
	"go.uber.org/zap"
)

func New(opts ...zap.Option) (l *zap.Logger, err error) {
	l, err = logConfig.Build(opts...)
	return
}

func NewLogger(opts ...zap.Option) *Logger {
	l, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return l.Sugar()
}

type Logger = zap.SugaredLogger
