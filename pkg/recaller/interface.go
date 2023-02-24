package recaller

import (
	"context"
)

type Recaller[t any] interface {
	SaveForReccal(ctx context.Context, key string, item t) error
	Stop()
	GetDeadChan() <-chan t
}

type Caller[t any] interface {
	Call(ctx context.Context, arg t) error
}
