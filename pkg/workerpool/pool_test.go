package workerpool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type j struct {
	num int
}

func (job j) Exec(ctx context.Context) int {
	time.Sleep(10 * time.Millisecond)
	return job.num
}

func TestPool(t *testing.T) {
	p := Create(0, 10, time.Second, WithRoundNRobin[int]())
	count := 1000

	m := make(map[int]struct{}, count)
	closed := false
	c := 0
	go func() {
		for res := range p.GetResult() {
			m[res] = struct{}{}
			c++
		}
		closed = true
	}()

	for i := 0; i < count; i++ {
		p.Put(j{i}, int64(i))
	}
	time.Sleep(1 * time.Second)
	p.Close()
	time.Sleep(time.Millisecond)
	assert.Equal(t, count, c)
	for i := 0; i < count; i++ {
		_, ok := m[i]
		assert.True(t, ok)
	}
	assert.True(t, closed)
}
