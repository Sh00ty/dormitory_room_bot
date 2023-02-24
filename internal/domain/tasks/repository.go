package tasks

import (
	"context"

	"github.com/Masterminds/squirrel"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

type Repository interface {
	Atomic(ctx context.Context, action func(ctx context.Context) error) error
	GetByID(ctx context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) (entities.Task, error)
	Create(ctx context.Context, task entities.Task, channelID valueObjects.ChannelID) error
	Delete(ctx context.Context, taskID valueObjects.TaskID, channelID valueObjects.ChannelID) error
	UpdateWorkers(ctx context.Context, task entities.Task) error
	UpdateNotifications(ctx context.Context, task entities.Task) error
	GetAll(ctx context.Context, conj squirrel.Sqlizer) (result []entities.Task, channelIDList []valueObjects.ChannelID, err error)
}
