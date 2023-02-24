package lists

import (
	"context"

	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

type Repository interface {
	GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.List, error)
	// list
	CreateList(ctx context.Context, channelID valueObjects.ChannelID, list entities.List) error
	GetList(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID) (entities.List, error)
	DeleteList(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID) error
	// items
	DeleteItem(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, index uint) error
	AddItem(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, item entities.Item) error
}
