package credits

import (
	"context"

	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

type Repository interface {
	Create(ctx context.Context, Credit *entities.Credit) error
	GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Credit, error)
	Get(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) (entities.Credit, error)
	Delete(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) error
	Update(ctx context.Context, Credit entities.Credit) error
	Atomic(ctx context.Context, action func(context.Context) error) error
}
