package sheduler

import "time"

type Scheduller interface {
	Start(interval time.Duration, action func() error) (finish func())
}
