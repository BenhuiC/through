package log

import (
	"go.uber.org/zap"
)

func New(opts ...zap.Option) (l *zap.Logger, err error) {
	l, err = logConfig.Build(opts...)
	return
}

func NewLogger(opts ...zap.Option) (*Logger, error) {
	l, err := New(opts...)
	if err != nil {
		return nil, err
	}
	return l.Sugar(), nil
}

type Logger = zap.SugaredLogger
