package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	repoIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/infra/repositories/tasks"
	uscIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/usecases/tasks"
	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	metric "github.com/Sh00ty/dormitory_room_bot/internal/metrics"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
	"go.uber.org/multierr"
)

type usecase struct {
	taskRepo repoIntf.Repository
}

func NewUsecase(r repoIntf.Repository) uscIntf.Usecase {
	return &usecase{taskRepo: r}
}

func (u *usecase) Create(ctx context.Context,
	taskID valueObjects.TaskID,
	typ entities.TaskType,
	title string,
	author string,
	workers []valueObjects.UserID,
	workerCount uint32,
	notifyInterval *time.Duration,
	notifyTime *time.Time,
	channelID valueObjects.ChannelID) error {

	task := entities.CreateTask(taskID, title, author, workers, workerCount, notifyInterval, notifyTime)

	switch typ {
	case entities.Subscribing:
		task.MakeSubscribed()
	case entities.OneShot:
		task.MakeOneShot()
	}

	err := u.taskRepo.Create(ctx, task, channelID)
	if err != nil {
		return fmt.Errorf("can't create task: %w", err)
	}
	metric.TotalSuccesfulPostgresRequests.WithLabelValues("Create").Inc()
	metric.TotalCreatedTasks.WithLabelValues(string(taskID)).Inc()
	return nil
}

func (u *usecase) Delete(ctx context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) error {
	err := u.taskRepo.Delete(ctx, taskID, channelID)
	if err != nil {
		return fmt.Errorf("Delete : %w", err)
	}
	metric.TotalSuccesfulPostgresRequests.WithLabelValues("delete").Inc()
	return nil
}

func (u *usecase) Get(ctx context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) (entities.Task, error) {
	task, err := u.taskRepo.GetByID(ctx, taskID, channelID)
	if err != nil {
		return entities.Task{}, fmt.Errorf("Get : %w", err)
	}
	metric.TotalSuccesfulPostgresRequests.WithLabelValues("GetByID").Inc()
	return task, nil
}

func (u *usecase) NotificationCheck(c context.Context) (result []entities.Task, reschannelIDList []valueObjects.ChannelID, err error) {

	err = u.taskRepo.Atomic(c, func(ctx context.Context) error {

		tasks, channelIDList, err2 := u.taskRepo.GetAll(ctx,
			squirrel.And{
				squirrel.LtOrEq{"notify_time": time.Now()},
				squirrel.NotEq{"notify_time": entities.NullDate},
			})

		if err2 != nil {
			return fmt.Errorf("NotificationCheck : %w", err2)
		}
		metric.TotalSuccesfulPostgresRequests.WithLabelValues("GetAll").Inc()

		for _, task := range tasks {
			task.UpdateNotificationDate()
			err2 = u.taskRepo.UpdateNotifications(ctx, task)
			if err2 != nil {
				logger.Errorf("NotificationCheck : can't update notification time : task_id: %d : %w", task.ID, err2)
				continue
			}
			metric.TotalSheduledTasks.WithLabelValues(string(task.ID)).Inc()
			metric.TotalSuccesfulPostgresRequests.WithLabelValues("NotificationCheck").Inc()
			result = append(result, task)
		}
		reschannelIDList = channelIDList
		return nil
	})

	if err != nil {
		err = multierr.Append(localerrors.ErrNothingToNotificate, err)
	}
	return
}

func (u *usecase) GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Task, error) {
	tasks, _, err := u.taskRepo.GetAll(ctx, squirrel.Eq{"channel_id": channelID})
	if err != nil {
		return nil, fmt.Errorf("GetAll : %w", err)
	}
	if len(tasks) == 0 {
		return nil, localerrors.ErrDoesntExist
	}
	metric.TotalSuccesfulPostgresRequests.WithLabelValues("GetAll").Inc()
	return tasks, nil
}

func (u *usecase) UpdateWorkers(c context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) (entities.Task, error) {
	resTask := entities.Task{}
	err := u.taskRepo.Atomic(c, func(ctx context.Context) error {
		task, err2 := u.taskRepo.GetByID(ctx, taskID, channelID)
		if err2 != nil {
			return fmt.Errorf("UpdateWorkers : %w", err2)
		}
		metric.TotalSuccesfulPostgresRequests.WithLabelValues("GetByID").Inc()
		err2 = task.Workers.GenerateNext()
		if err2 != nil {
			return fmt.Errorf("UpdateWorkers : no workers : %w", err2)
		}

		err2 = u.taskRepo.UpdateWorkers(ctx, task)
		if err2 != nil {
			return fmt.Errorf("UpdateWorkers : can't update workers : %w", err2)
		}
		metric.TotalSuccesfulPostgresRequests.WithLabelValues("UpdateWorkers").Inc()

		resTask = task
		return nil
	})

	if err != nil {
		err = multierr.Append(localerrors.ErrCantUpdateWorkers, err)
	}
	return resTask, err
}

func (u *usecase) Subscribe(c context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID, worker valueObjects.UserID) error {
	err := u.taskRepo.Atomic(c, func(ctx context.Context) error {
		task, err2 := u.taskRepo.GetByID(ctx, taskID, channelID)
		if err2 != nil {
			return fmt.Errorf("Subscribe : %w", err2)
		}
		metric.TotalSuccesfulPostgresRequests.WithLabelValues("GetByID").Inc()

		if !task.IsSubscribed() {
			return fmt.Errorf("Subscribe : %w", localerrors.ErrNotSubTask)
		}

		err2 = task.Sub(worker)
		if err2 != nil {
			return fmt.Errorf("Subscribe : %w", err2)
		}
		err2 = u.taskRepo.UpdateWorkers(ctx, task)
		if err2 != nil {
			return fmt.Errorf("Subscribe : can't update workers : %w", err2)
		}
		metric.TotalSuccesfulPostgresRequests.WithLabelValues("Subscribe").Inc()
		return nil
	})

	if err != nil {
		err = multierr.Append(localerrors.ErrCantUpdateWorkers, err)
	}
	return err
}

func (u *usecase) UnSubscribe(c context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID, worker valueObjects.UserID) error {
	err := u.taskRepo.Atomic(c, func(ctx context.Context) error {
		task, err2 := u.taskRepo.GetByID(ctx, taskID, channelID)
		if err2 != nil {
			return fmt.Errorf("UpdateWorkers : %w", err2)
		}
		metric.TotalSuccesfulPostgresRequests.WithLabelValues("GetByID").Inc()

		if !task.IsSubscribed() {
			return fmt.Errorf("UnSubscribe : %w", localerrors.ErrNotSubTask)
		}

		err2 = task.Unsub(worker)
		if err2 != nil {
			return fmt.Errorf("UpdateWorkers : no workers : %w", err2)
		}
		err2 = u.taskRepo.UpdateWorkers(ctx, task)
		if err2 != nil {
			return fmt.Errorf("UpdateWorkers : can't update workers : %w", err2)
		}
		metric.TotalSuccesfulPostgresRequests.WithLabelValues("UpdateWorkers").Inc()
		return nil
	})

	if err != nil {
		err = multierr.Append(localerrors.ErrCantUpdateWorkers, err)
	}
	return err
}
