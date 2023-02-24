package credits

import (
	"context"

	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

type CreditManager interface {
	Create(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID, credit valueObjects.Money) error
	GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Credit, error)
	CreditTransaction(c context.Context, channelID valueObjects.ChannelID, buyer valueObjects.UserID, payers []valueObjects.UserID, credit valueObjects.Money) error
	Checkout(c context.Context, channelID valueObjects.ChannelID) ([]entities.Transaction, error)
}

type Resolver interface {
	ResolveCredits(data []entities.Credit) []entities.Transaction
}
