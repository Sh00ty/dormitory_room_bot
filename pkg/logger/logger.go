package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	})

	errorFatalLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.ErrorLevel || level == zapcore.FatalLevel
	})
	stdoutSyncer := zapcore.Lock(os.Stdout)
	stderrSyncer := zapcore.Lock(os.Stderr)

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
			stdoutSyncer,
			infoLevel,
		),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
			stderrSyncer,
			errorFatalLevel,
		),
	)
	logger := zap.New(core)
	logger.Named("mlogger")
	zap.ReplaceGlobals(logger)
}

func GetStat() []string {
	return nil
}

func Info(message string) {
	zap.L().Info(message)
}
func Infof(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	zap.L().Info(message)
}
func Debug(message string) {
	zap.L().Debug(message)
}
func Debugf(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	zap.L().Debug(message)
}
func Error(message string) {
	zap.L().Error(message)
}
func Errorf(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	zap.L().Error(message)
}

func Fataf(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	zap.L().Fatal(message)
}
