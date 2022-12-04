package user

import (
	"context"
	"fmt"

	repoIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/infra/repositories/users"
	uscIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/usecases/users"
	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	metrics "github.com/Sh00ty/dormitory_room_bot/internal/metrics"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
	"github.com/prometheus/client_golang/prometheus"
)

type usecase struct {
	repo repoIntf.Repository
}

func New(r repoIntf.Repository) uscIntf.UserService {
	return &usecase{
		repo: r,
	}
}

func (u usecase) GetBatchFromIDs(ctx context.Context, userIDList []valueObjects.UserID) ([]entities.User, error) {
	users, err := u.repo.GetBatch(ctx, userIDList)
	if err != nil {
		return nil, fmt.Errorf("GetBatch : %w", err)
	}

	metrics.TotalSuccesfulPostgresRequests.With(prometheus.Labels{"request": "GetBatchFromIDs"}).Inc()
	if len(users) == 0 {
		return nil, fmt.Errorf("GetBatch : %w", localerrors.ErrDoesntExist)
	}
	return users, nil
}

func (u usecase) GetBatchFromUsernames(ctx context.Context, usernameList []string) ([]entities.User, error) {
	users, err := u.repo.GetBatchFromUsernames(ctx, usernameList)
	if err != nil {
		return nil, fmt.Errorf("GetBatchFromUsernames : %w", err)
	}

	metrics.TotalSuccesfulPostgresRequests.With(prometheus.Labels{"request": "GetBatchFromUsernames"}).Inc()
	if len(users) == 0 {
		return nil, fmt.Errorf("GetBatchFromUsernames : %w", localerrors.ErrDoesntExist)
	}
	return users, nil
}

func (u usecase) Create(ctx context.Context, usr entities.User) error {
	err := u.repo.Create(ctx, usr)
	if err != nil {
		return fmt.Errorf("Create : %w", err)
	}
	metrics.TotalSuccesfulPostgresRequests.With(prometheus.Labels{"request": "create"}).Inc()
	return nil
}

func (u usecase) Delete(ctx context.Context, userID valueObjects.UserID) error {
	err := u.repo.Delete(ctx, userID)
	if err != nil {
		return fmt.Errorf("Delete : %w", err)
	}
	metrics.TotalSuccesfulPostgresRequests.With(prometheus.Labels{"request": "delete"}).Inc()
	return nil
}
