package recaller

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/logger"
)

type Repository[t any] interface {
	SaveForReccal(ctx context.Context, item Item[t]) error
	GetItemsForResend(ctx context.Context) ([]Item[t], error)
	DeleteItems(ctx context.Context, items []Item[t]) error
	UpdateItems(ctx context.Context, items []Item[t]) error
	Delete(ctx context.Context)
}

type Item[t any] struct {
	Reccals        uint
	Key            string
	NextReccalTime time.Time
	Args           t
}

func (i Item[t]) MarshalBinary() (data []byte, err error) {
	return json.Marshal(i)
}
func (i *Item[t]) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, i)
}

type recaller[t any] struct {
	cache          Repository[t]
	timeout        time.Duration
	maxReccalCount uint
	deadChan       chan t
	withDeadChan   bool
	closeChan      chan struct{}
	srtategy       regrassionStrategy
	baseTimeout    time.Duration
	caller         Caller[t]
}

type regrassionStrategy int

const (
	evryCron = iota
	linear
	exponential
)

func (r recaller[t]) Stop() {
	if r.withDeadChan {
		close(r.closeChan)
		close(r.deadChan)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r.cache.Delete(ctx)
}

func Init[t any](ctx context.Context, caller Caller[t], timeout, baseTimeout time.Duration, maxReccalCount uint, opts ...Opt[t]) (Recaller[t], error) {
	r := recaller[t]{
		timeout:        timeout,
		baseTimeout:    baseTimeout,
		maxReccalCount: maxReccalCount,
		caller:         caller,
	}

	for _, opt := range opts {
		opt(&r)
	}

	go func() {
		ctx := context.Background()
		for {
			select {
			case <-r.closeChan:
				return
			case <-time.After(r.timeout):
				items, err := r.cache.GetItemsForResend(ctx)
				if err != nil {
					logger.Errorf("can't get items for resend %v", err)
					continue
				}

				itemsToDelete := make([]Item[t], 0, len(items))
				itemsToUpdate := make([]Item[t], 0, len(items))
				for _, item := range items {
					err := r.caller.Call(ctx, item.Args)
					if err != nil {
						logger.Errorf("can't recall item %s due to err: %v", item.Key, err)
						if item.Reccals >= r.maxReccalCount {
							if r.withDeadChan {
								go r.sendToDeadChan(item.Args)
							}
							itemsToDelete = append(itemsToDelete, item)
							continue
						}
						itemsToUpdate = append(itemsToUpdate, r.updateNextReccalTime(item))
						continue
					}
					logger.Infof("sucessfuly recalled item %s", item.Key)
					itemsToDelete = append(itemsToDelete, item)
				}
				r.updateItems(ctx, itemsToUpdate)
				r.deleteItems(ctx, itemsToDelete)
			}

		}
	}()

	return r, nil
}

func (r recaller[t]) sendToDeadChan(item t) {
	select {
	case <-r.closeChan:
		return
	case <-time.After(5 * time.Second):
		logger.Errorf("can't send item to dead channel for a long time")
		return
	case r.deadChan <- item:
		return
	}
}

func (r recaller[t]) updateNextReccalTime(item Item[t]) Item[t] {

	addition := time.Duration(0)
	switch r.srtategy {
	case linear:
		addition = time.Duration(item.Reccals) * r.baseTimeout
	case exponential:
		addition = time.Duration(math.Exp(float64(item.Reccals))) * r.baseTimeout
	default:
		addition = 0
	}
	item.Reccals++
	item.NextReccalTime = time.Now().Add(addition)
	return item
}

func (r recaller[t]) deleteItems(ctx context.Context, itemsToDelete []Item[t]) {
	if len(itemsToDelete) != 0 {
		err := r.cache.DeleteItems(ctx, itemsToDelete)
		if err != nil {
			logger.Errorf("can't delete items from recaller cache: %v", err)
		}
	}
}

func (r recaller[t]) updateItems(ctx context.Context, itemsToUpdate []Item[t]) {
	if len(itemsToUpdate) != 0 {
		err := r.cache.UpdateItems(ctx, itemsToUpdate)
		if err != nil {
			logger.Errorf("can't update items from recaller cache: %v", err)
		}
	}
}

func (r recaller[t]) SaveForReccal(ctx context.Context, key string, args t) error {
	return r.cache.SaveForReccal(ctx, Item[t]{
		NextReccalTime: time.Now(),
		Args:           args,
		Key:            key,
	})
}

func (r recaller[t]) GetDeadChan() <-chan t {
	return r.deadChan
}
