package credits

import (
	"context"

	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type UseCaseInterface interface {
	Create(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID, credit valueObjects.Money) error
	GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Credit, error)
	Get(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) (entities.Credit, error)
	Delete(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) error
	CreditTransaction(ctx context.Context, channelID valueObjects.ChannelID, buyer valueObjects.UserID, users []valueObjects.UserID, credit valueObjects.Money) error
	ClearBalances(ctx context.Context, channelID valueObjects.ChannelID) error
}

type Resolver interface {
	ResolveCredits(data []entities.Credit) []entities.Transaction
}
