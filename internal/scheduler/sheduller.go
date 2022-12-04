package sheduler

import (
	"runtime/debug"
	"time"

	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
)

type sheduller struct {
	interval time.Duration
	action   func() error
	done     chan struct{}
}

func NewTimeSheduller() Scheduller {
	return &sheduller{}
}

func (s *sheduller) Start(interval time.Duration, action func() error) (finish func()) {
	s.interval = interval
	s.action = action
	s.done = make(chan struct{})
	go s.run()
	return func() {
		s.done <- struct{}{}
	}
}

func (s sheduller) run() {
	for {
		select {
		case <-s.done:
			close(s.done)
			return
		case <-time.After(s.interval):
			err := func() error {
				defer func() {
					if err := recover(); err != nil {
						logger.Errorf("PANIC PANIC PANIC : %v : %s", err, debug.Stack())
					}
				}()
				err2 := s.action()
				return err2
			}()
			if err != nil {
				logger.Debugf("time sheduler : %v", err)
			}
		}
	}

}
