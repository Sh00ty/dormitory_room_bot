package recaller

import (
	"context"
	"encoding/json"
	"time"
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
