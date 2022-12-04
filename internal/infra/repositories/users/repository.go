package user

import (
	"context"

	"github.com/Masterminds/squirrel"
	repoIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/infra/repositories/users"
	pgxbalancer "github.com/Sh00ty/dormitory_room_bot/internal/transaction_balancer"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	templateRepo "github.com/Sh00ty/dormitory_room_bot/pkg/pg_generick_repo"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

const dbName = "users"

type repository struct {
	pgxbalancer.TransactionBalancer
}

func NewRepo(balancer pgxbalancer.TransactionBalancer) repoIntf.Repository {
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
