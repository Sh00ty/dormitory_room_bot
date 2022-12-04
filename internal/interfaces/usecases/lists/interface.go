package lists

import (
	"context"

	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type ListManager interface {
	GetAllChannelLists(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.List, error)

	// list

	CreateList(ctx context.Context, channelID valueObjects.ChannelID, list entities.List) error

	GetList(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID) (entities.List, error)

	DeleteList(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID) error

	// items

	GetRandomItems(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, count uint64) ([]entities.Item, error)

	DeleteByIndex(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, index uint) error

	AddItem(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, item entities.Item) error
}
