package recaller

import (
	"context"
	"encoding/json"
	"time"

	repoIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/infra/repositories/recaller"
	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	"github.com/go-redis/redis/v8"
)

type repository[t any] struct {
	client  redis.UniversalClient
	setName string
}

func NewRedisRepo[t any](addr, password, setName string) repoIntf.Repository[t] {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp := client.Ping(ctx)
	if resp.Err() != nil {
		panic(resp.Err())
	}
	return repository[t]{
		client:  client,
		setName: setName,
	}
}

func (r repository[t]) SaveForReccal(ctx context.Context, item repoIntf.Item[t]) error {
	return r.client.HSet(ctx, r.setName, item.Key, item).Err()
}

func (r repository[t]) UpdateItems(ctx context.Context, items []repoIntf.Item[t]) (err error) {
	for _, item := range items {
		err = r.client.HSet(ctx, r.setName, item.Key, item).Err()
		if err != nil {
			logger.Errorf("failed to update item %v: %v", item.Key, err)
		}
	}
	return err
}

func (r repository[t]) GetItemsForResend(ctx context.Context) ([]repoIntf.Item[t], error) {
	result := r.client.HGetAll(ctx, r.setName)
	if result.Err() != nil {
		return nil, result.Err()
	}

	res := make([]repoIntf.Item[t], 0, 10)
	itemMap, err := result.Result()
	if err != nil {
		return nil, err
	}

	for key, item := range itemMap {
		it := repoIntf.Item[t]{}
		err = json.Unmarshal([]byte(item), &it)
		if err != nil {
			logger.Errorf("can't unmarshal item %v: %v", key, err)
			continue
		}
		if it.NextReccalTime.After(time.Now()) {
			continue
		}
		res = append(res, it)
	}

	return res, nil
}

func (r repository[t]) DeleteItems(ctx context.Context, items []repoIntf.Item[t]) error {
	itd := make([]string, 0, len(items))
	for _, item := range items {
		itd = append(itd, item.Key)
	}
	return r.client.HDel(ctx, r.setName, itd...).Err()
}

func (r repository[t]) Delete(ctx context.Context) {
	err := r.client.Close()
	if err != nil {
		logger.Errorf("can't close redis client: %v", err)
	}
}
