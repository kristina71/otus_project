package logger

import (
	"fmt"

	"github.com/kristina71/otus_project/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(loggerConfig config.LoggerConfig) error {
	cnf := zap.NewProductionConfig()

	cnf.EncoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	cnf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cnf.ErrorOutputPaths = []string{"stderr"}
	cnf.OutputPaths = []string{loggerConfig.File, "stdout"}
	if err := cnf.Level.UnmarshalText([]byte(loggerConfig.Level)); err != nil {
		return fmt.Errorf("error building logger, can't parse logger level from file: %w", err)
	}
	currentLogger, err := cnf.Build()
	if err != nil {
		return fmt.Errorf("error building logger: %w", err)
	}
	zap.ReplaceGlobals(currentLogger)
	return nil
}
