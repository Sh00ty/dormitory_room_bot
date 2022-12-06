package pool

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	metric "github.com/Sh00ty/dormitory_room_bot/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type workerPool[resType any] struct {
	resChan        chan resType
	workers        []chan Job[resType]
	closers        chan struct{}
	workerCount    int64
	roundNrobinPtr int64
	srtategy       putStrategy
}

type putStrategy int8

const (
	hash        = 0
	roundNrobin = 1
)

func CreateWorkerPool[resType any](bufferSize uint64, concurrensy uint64, workerResultTimeOut time.Duration, opts ...Opt[resType]) Pool[resType] {

	if concurrensy == 0 {
		panic("worker pool can't be with comcurrensy = 0")
	}

	wp := &workerPool[resType]{
		workers:     make([]chan Job[resType], 0, concurrensy),
		closers:     make(chan struct{}),
		resChan:     make(chan resType, bufferSize),
		workerCount: int64(concurrensy),
	}

	for _, opt := range opts {
		opt(wp)
	}

	for i := uint64(0); i < concurrensy; i++ {
		wp.workers = append(wp.workers, make(chan Job[resType], bufferSize))
		go func(k uint64) {
			defer func() {
				if err := recover(); err != nil {
					logger.Errorf("PANIC:PANIC:PANIC : %v; w=%d\n\n%s", err, k, string(debug.Stack()))
				}
			}()

			for {
				select {
				case job := <-wp.workers[k]:
					ctx, cancel := context.WithTimeout(context.Background(), workerResultTimeOut)
					res := job.Exec(ctx)
					select {
					case <-wp.closers:
						cancel()
						return
					case <-ctx.Done():
						logger.Errorf("worker %d dedline exedeed", k)
						cancel()
					case wp.resChan <- res:
						cancel()
					}
				case <-wp.closers:
					return
				}
			}
		}(i)
	}
	return wp
}

func (wp *workerPool[resType]) Close() {
	close(wp.closers)
	for _, workerChan := range wp.workers {
		close(workerChan)
	}
	close(wp.resChan)
}

func (wp *workerPool[resType]) GetResult() <-chan resType {
	return wp.resChan
}

func (wp *workerPool[resType]) Put(j Job[resType], hashID int64) int64 {
	timer := prometheus.NewTimer(metric.PutExecutingTime)
	defer timer.ObserveDuration().Milliseconds()

	workerInd := int64(0)
	switch wp.srtategy {
	case hash:
		if hashID < 0 {
			hashID = -hashID
		}
		workerInd = hashID % wp.workerCount
	case roundNrobin:
		wp.roundNrobinPtr = (wp.roundNrobinPtr + 1) % wp.workerCount
		workerInd = wp.roundNrobinPtr
	}

	wp.workers[workerInd] <- j
	return workerInd
}

type Opt[resType any] func(*workerPool[resType])

func WithRoundNRobin[resType any]() Opt[resType] {
	return func(wp *workerPool[resType]) {
		wp.srtategy = roundNrobin
	}
}
