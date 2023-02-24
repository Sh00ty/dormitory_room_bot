package user

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	metrics "gitlab.com/Sh00ty/dormitory_room_bot/internal/metrics"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

type usecase struct {
	repo Repository
}

func New(r Repository) *usecase {
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
