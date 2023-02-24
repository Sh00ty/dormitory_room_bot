package user

import (
	"context"

	"github.com/Masterminds/squirrel"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
	templateRepo "gitlab.com/Sh00ty/dormitory_room_bot/pkg/pggen"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/pgxbalancer"
)

const dbName = "users"

type repository struct {
	pgxbalancer.TransactionBalancer
}

func NewRepo(balancer pgxbalancer.TransactionBalancer) repository {
	return repository{TransactionBalancer: balancer}
}

func (r repository) GetRunnner(ctx context.Context) pgxbalancer.Runner {
	return r.TransactionBalancer.GetRunnner(ctx)
}

func (r repository) TableName() string {
	return dbName
}

func (r repository) GetBatch(ctx context.Context, userIDList []valueObjects.UserID) ([]entities.User, error) {
	userDtoList, err := templateRepo.GetAllBy[userDTO](ctx, r, squirrel.Eq{"id": userIDList})
	if err != nil {
		return nil, err
	}
	return UserListFromDTOList(userDtoList), nil
}

func (r repository) GetBatchFromUsernames(ctx context.Context, usernameList []string) ([]entities.User, error) {
	userDtoList, err := templateRepo.GetAllBy[userDTO](ctx, r, squirrel.Eq{"username": usernameList})
	if err != nil {
		return nil, err
	}
	return UserListFromDTOList(userDtoList), nil
}

func (r repository) Delete(ctx context.Context, userID valueObjects.UserID) error {
	return templateRepo.Delete[userDTO](ctx, r, squirrel.Eq{"id": userID})
}
func (r repository) Create(ctx context.Context, usr entities.User) error {
	return templateRepo.Create(ctx, r, userDTO{
		ID:          usr.ID,
		Username:    usr.UserName,
		PhoneNumber: usr.PhoneNumber,
	})
}
