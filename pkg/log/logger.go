package log

import (
	"go.uber.org/zap"
)

var sugar = newSugar()

func newSugar() *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return l.WithOptions(zap.AddStacktrace(zap.ErrorLevel)).Sugar()
}

func Zap() *zap.SugaredLogger {
	return newSugar()
}

const (
	Source = "source"
	Stack  = "stack"
)

func WithSource(logger *zap.Logger, source string) *zap.Logger {
	return logger.With(zap.String(Source, source))
}

func WithStack(logger *zap.Logger) *zap.Logger {
	return logger.With(zap.Stack(Stack))
}
