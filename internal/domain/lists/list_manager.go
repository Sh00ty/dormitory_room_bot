package lists

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

type listManager struct {
	repo Repository
}

func NewListManager(repo Repository) *listManager {
	return &listManager{repo: repo}
}

func (l *listManager) GetAllChannelLists(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.List, error) {
	lists, err := l.repo.GetAll(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("GetAllChannelLists : %w", err)
	}
	return lists, nil
}

// list

func (l *listManager) CreateList(ctx context.Context, channelID valueObjects.ChannelID, list entities.List) error {
	err := l.repo.CreateList(ctx, channelID, list)
	if err != nil {
		return fmt.Errorf("CreateList : %w", err)
	}
	return nil
}

func (l *listManager) GetList(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID) (entities.List, error) {
	list, err := l.repo.GetList(ctx, channelID, listID)
	if err != nil {
		return entities.List{}, fmt.Errorf("GetList : %w", err)
	}
	return list, nil
}

func (l *listManager) GetRandomItems(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, count uint64) ([]entities.Item, error) {
	list, err := l.repo.GetList(ctx, channelID, listID)
	if err != nil {
		return nil, fmt.Errorf("GetRandomItems : %w", err)
	}

	itemCount := uint64(len(list.Items))
	if itemCount == 0 {
		return nil, localerrors.ErrNoItems
	}

	if itemCount < count {
		count = itemCount
	}

	res := make([]entities.Item, 0, count)

	for i := uint64(0); i < count; i++ {
		ind, err := rand.Int(rand.Reader, big.NewInt(int64(itemCount)))
		if err != nil {
			return nil, err
		}
		res = append(res, list.Items[ind.Int64()])
		list.Items = append(list.Items[:ind.Int64()], list.Items[ind.Int64()+1:]...)
		itemCount--
	}

	return res, nil
}

func (l *listManager) DeleteList(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID) error {
	err := l.repo.DeleteList(ctx, channelID, listID)
	if err != nil {
		return fmt.Errorf("DeleteList : %w", err)
	}
	return nil
}

func (l *listManager) DeleteByIndex(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, index uint) error {
	err := l.repo.DeleteItem(ctx, channelID, listID, index)
	if err != nil {
		return fmt.Errorf("DeleteByIndex : %w", err)
	}
	return nil

}

func (l *listManager) AddItem(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, item entities.Item) error {
	err := l.repo.AddItem(ctx, channelID, listID, item)
	if err != nil {
		return fmt.Errorf("AddItem : %w", err)
	}
	return nil
}
