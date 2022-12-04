package lists

import (
	"context"
	"fmt"
	mrand "math/rand"
	"time"

	repoIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/infra/repositories/lists"
	uscIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/usecases/lists"
	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type listManager struct {
	repo repoIntf.Repository
}

func NewListManager(repo repoIntf.Repository) uscIntf.ListManager {
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
		mrand.Seed(time.Now().UnixNano())
		ind := mrand.Uint64() % uint64(itemCount)
		res = append(res, list.Items[ind])
		list.Items = append(list.Items[:ind], list.Items[ind+1:]...)
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
