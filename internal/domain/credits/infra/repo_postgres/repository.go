package credits

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
	templateRepo "gitlab.com/Sh00ty/dormitory_room_bot/pkg/pggen"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/pgxbalancer"
)

const dbName = "credits"

type creditRepository struct {
	pgxbalancer.TransactionBalancer
}

func NewDebitRepository(balancer pgxbalancer.TransactionBalancer) *creditRepository {
	return &creditRepository{
		balancer,
	}
}

func (r creditRepository) TableName() string {
	return dbName
}

func (r creditRepository) GetRunnner(ctx context.Context) pgxbalancer.Runner {
	return r.TransactionBalancer.GetRunnner(ctx)
}

func (r *creditRepository) Create(ctx context.Context, Credit *entities.Credit) error {
	return templateRepo.Create(ctx, r, *Credit)
}

func (r *creditRepository) GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Credit, error) {
	return templateRepo.GetAllBy[entities.Credit](ctx, r, squirrel.Eq{"channel_id": channelID})
}

func (r *creditRepository) Get(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) (entities.Credit, error) {
	return templateRepo.GetBy[entities.Credit](ctx, r, squirrel.Eq{"channel_id": channelID, "user_id": userID})
}

func (r *creditRepository) Delete(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) error {
	return templateRepo.Delete[entities.Credit](ctx, r, squirrel.Eq{"channel_id": channelID, "user_id": userID})
}

func (r *creditRepository) Update(ctx context.Context, Credit entities.Credit) error {
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
