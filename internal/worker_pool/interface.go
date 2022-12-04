package pool

import "context"

type Job[resType any] interface {
	Exec(ctx context.Context) resType
}
type Pool[resType any] interface {
	Put(j Job[resType], hashID int64) int64
	Close()
	GetResult() <-chan resType
}
