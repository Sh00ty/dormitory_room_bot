package credits

import (
	"context"

	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type Repository interface {
	Create(ctx context.Context, Credit *entities.Credit) error
	GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Credit, error)
	Get(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) (entities.Credit, error)
	Delete(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) error
	Update(ctx context.Context, Credit entities.Credit) error
	Atomic(ctx context.Context, action func(context.Context) error) error
}
