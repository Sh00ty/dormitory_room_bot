package tasks

import (
	"context"
	"time"

	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

type TaskManager interface {
	Create(ctx context.Context, taskID valueObjects.TaskID, typ entities.TaskType, title, author string, workers []valueObjects.UserID, workerCount uint32, notifyInterval *time.Duration, notifyTime *time.Time, channelID valueObjects.ChannelID) error
	Delete(ctx context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) error
	Get(ctx context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) (entities.Task, error)
	NotificationCheck(ctx context.Context) ([]entities.Task, []valueObjects.ChannelID, error)
	GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Task, error)
	UpdateWorkers(ctx context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) (entities.Task, error)
	Subscribe(c context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID, worker valueObjects.UserID) error
	UnSubscribe(c context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID, worker valueObjects.UserID) error
}
