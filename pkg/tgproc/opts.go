package tgproc

import (
	"time"

	pool "gitlab.com/Sh00ty/dormitory_room_bot/pkg/workerpool"
)

func WithResends(resendLimit uint, resendBaseTimeout time.Duration) Option {
	return func(b *Bot) {
		b.resendLimit = resendLimit
		b.resendBaseTimeout = resendBaseTimeout
	}
}

func WithResend(resendLimit uint, resendBaseTimeout time.Duration) Option {
	return func(b *Bot) {
		b.resendLimit = resendLimit
		b.resendBaseTimeout = resendBaseTimeout
	}
}

func WithLogger(l BotLogger) Option {
	return func(b *Bot) {
		b.logger = l
	}
}

func WithWorlerPool(concurrency, channelBufferSize uint64, commandExecTimeout time.Duration) Option {
	return func(b *Bot) {
		b.pool = pool.Create(
			channelBufferSize,
			concurrency,
			commandExecTimeout,
			pool.WithRoundNRobin[Messages](),
		)
	}
}
