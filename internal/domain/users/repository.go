package user

import (
	"context"

	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

type Repository interface {
	GetBatch(ctx context.Context, userIDList []valueObjects.UserID) ([]entities.User, error)
	GetBatchFromUsernames(ctx context.Context, usernameList []string) ([]entities.User, error)
	Delete(ctx context.Context, userID valueObjects.UserID) error
	Create(ctx context.Context, usr entities.User) error
}
