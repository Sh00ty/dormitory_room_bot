package tgproc

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type logger struct {
	zerolog.Logger
}

func (l logger) Errorf(message string, args ...interface{}) {
	l.Error().Time("time", time.Now()).Str("service", "tgframe").Msgf(message, args...)
}

func newLogger() *logger {
	return &logger{
		zerolog.New(os.Stderr),
	}
}
