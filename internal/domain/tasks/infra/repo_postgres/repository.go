package tasks

import (
	"context"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
	templateRepo "gitlab.com/Sh00ty/dormitory_room_bot/pkg/pggen"
	pgxbalancer "gitlab.com/Sh00ty/dormitory_room_bot/pkg/pgxbalancer"
)

const dbName = "tasks"

type repository struct {
	pgxbalancer.TransactionBalancer
}

func (r repository) TableName() string {
	return dbName
}

func (r repository) GetRunnner(ctx context.Context) pgxbalancer.Runner {
	return r.TransactionBalancer.GetRunnner(ctx)
}

func NewRepo(balancer pgxbalancer.TransactionBalancer) *repository {
	return &repository{balancer}
}

func (r repository) GetByID(ctx context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) (entities.Task, error) {
	dto, err := templateRepo.GetBy[taskDTO](ctx, r, squirrel.Eq{"id": taskID, "channel_id": channelID})
	if err != nil {
		return entities.Task{}, err
	}
	return TaskFromDTO(dto), nil
}

func (r repository) Create(ctx context.Context, task entities.Task, channelID valueObjects.ChannelID) error {
	notifyTime, notifyInterval := getNotificatorValues(task)
	workers, workerCount, firstWorkerPtr := getWokerPermutationValues(task)

	converdetWorkers := workers.([]valueObjects.UserID)
	strWorker := make([]string, 0, len(converdetWorkers))

	for _, worker := range converdetWorkers {
		strWorker = append(strWorker, string(worker))
	}

	dto := taskDTO{
		ID:             task.ID,
		Title:          task.Title,
		NotifyTime:     notifyTime.(time.Time),
		NotifyInterval: notifyInterval.(time.Duration),
		Workers:        strWorker,
		WorkerCount:    workerCount.(uint32),
		FirstWorker:    firstWorkerPtr.(uint32),
		ChannelID:      channelID,
		Stage:          task.Sstage,
		Type:           task.Type,
		Author:         task.Author,
	}

	return templateRepo.Create(ctx, r, dto)
}

func (r repository) Delete(ctx context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) error {
	return templateRepo.Delete[taskDTO](ctx, r, squirrel.Eq{"id": taskID, "channel_id": channelID})
}

func (r repository) UpdateNotifications(ctx context.Context, task entities.Task) (err error) {
	notifyTime, notifyInterval := getNotificatorValues(task)
	sql, values, err := squirrel.Update(dbName).
		Set("notify_time", notifyTime).
		Set("notify_interval", notifyInterval).
		Where(squirrel.Eq{"id": task.ID}).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}

	commandTag, err := r.TransactionBalancer.GetRunnner(ctx).Exec(ctx, sql, values...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return localerrors.ErrDoesntExist
		}
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return localerrors.ErrDidntUpdated
	}
	return nil
}

func (r repository) UpdateWorkers(ctx context.Context, task entities.Task) (err error) {
	workers, workerCount, firstWorkerPtr := getWokerPermutationValues(task)
	sql, values, err := squirrel.Update(dbName).
		Set("workers", workers).
		Set("worker_count", workerCount).
		Set("first_worker", firstWorkerPtr).
		Where(squirrel.Eq{"id": task.ID}).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}

	commandTag, err := r.TransactionBalancer.GetRunnner(ctx).Exec(ctx, sql, values...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return localerrors.ErrDoesntExist
		}
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return localerrors.ErrDidntUpdated
	}
	return nil
}

func (r repository) GetAll(ctx context.Context, conj squirrel.Sqlizer) (result []entities.Task, channelIDList []valueObjects.ChannelID, err error) {
	dtoList, err := templateRepo.GetAllBy[taskDTO](ctx, r, conj)
	if err != nil {
		return nil, nil, err
	}
	for _, dto := range dtoList {
		channelIDList = append(channelIDList, valueObjects.ChannelID(dto.ChannelID))
		result = append(result, TaskFromDTO(dto))
	}
	return
}
