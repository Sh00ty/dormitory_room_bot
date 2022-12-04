package repository

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	pgxbalancer "github.com/Sh00ty/dormitory_room_bot/internal/transaction_balancer"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	templateRepo "github.com/Sh00ty/dormitory_room_bot/pkg/pg_generick_repo"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
	"github.com/jackc/pgx/v4"
)

const dbName = "credits"

type DebitRepository struct {
	pgxbalancer.TransactionBalancer
}

func NewDebitRepository(balancer pgxbalancer.TransactionBalancer) *DebitRepository {
	return &DebitRepository{
		balancer,
	}
}

func (r DebitRepository) TableName() string {
	return dbName
}

func (r DebitRepository) GetRunnner(ctx context.Context) pgxbalancer.Runner {
	return r.TransactionBalancer.GetRunnner(ctx)
}

func (r *DebitRepository) Create(ctx context.Context, Credit *entities.Credit) error {
	return templateRepo.Create(ctx, r, *Credit)
}

func (r *DebitRepository) GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Credit, error) {
	return templateRepo.GetAllBy[entities.Credit](ctx, r, squirrel.Eq{"channel_id": channelID})
}

func (r *DebitRepository) Get(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) (entities.Credit, error) {
	return templateRepo.GetBy[entities.Credit](ctx, r, squirrel.Eq{"channel_id": channelID, "user_id": userID})
}

func (r *DebitRepository) Delete(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) error {
	return templateRepo.Delete[entities.Credit](ctx, r, squirrel.Eq{"channel_id": channelID, "user_id": userID})
}

func (r *DebitRepository) Update(ctx context.Context, Credit entities.Credit) error {
	sql, values, err := squirrel.Update(r.TableName()).
		Set("credit", Credit.Credit).
		Where(squirrel.Eq{"channel_id": Credit.ChannelID, "user_id": Credit.UserID}).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}

	commandTag, err := r.TransactionBalancer.GetRunnner(ctx).Exec(ctx, sql, values...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return localerrors.ErrDoesntExist
		}
	}

	if commandTag.RowsAffected() == 0 {
		return localerrors.ErrDidntUpdated
	}

	return nil
}
