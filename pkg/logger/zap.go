package logger

import "go.uber.org/zap"

type zapLogger struct {
	*zap.Logger
}

func NewZapLogger(isDev bool) Logger {
	var log *zap.Logger
	var err error
	if isDev {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	return &zapLogger{log}
}

func (z *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{z.Logger.With(fields...)}
}
