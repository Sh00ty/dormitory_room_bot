package user

import (
	"context"

	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type Repository interface {
	GetBatch(ctx context.Context, userIDList []valueObjects.UserID) ([]entities.User, error)
	GetBatchFromUsernames(ctx context.Context, usernameList []string) ([]entities.User, error)
	Delete(ctx context.Context, userID valueObjects.UserID) error
	Create(ctx context.Context, usr entities.User) error
}
