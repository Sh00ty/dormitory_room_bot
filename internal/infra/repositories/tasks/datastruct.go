package tasks

import (
	"reflect"
	"time"

	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type taskDTO struct {
	ID             valueObjects.TaskID       `db:"id"`
	Title          string                    `db:"title"`
	NotifyTime     time.Time                 `db:"notify_time"`
	NotifyInterval time.Duration             `db:"notify_interval"`
	Workers        []string                  `db:"workers"`
	WorkerCount    uint32                    `db:"worker_count"`
	FirstWorker    uint32                    `db:"first_worker"`
	ChannelID      valueObjects.ChannelID    `db:"channel_id"`
	Author         string                    `db:"author"`
	Type           entities.TaskType         `db:"type"`
	Stage          entities.SubscribedStages `db:"stage"`
}

func TaskFromDTO(dto taskDTO) entities.Task {
	var task entities.Task
	task.Notificator = entities.CreateNotificator(&dto.NotifyInterval, &dto.NotifyTime)
	task.Title = dto.Title
	task.ID = valueObjects.TaskID(dto.ID)
	task.Type = dto.Type
	task.Author = dto.Author
	task.Sstage = dto.Stage

	workers := make([]valueObjects.UserID, len(dto.Workers))
	for i := range dto.Workers {
		workers[i] = valueObjects.UserID(dto.Workers[i])
	}
	task.Workers = entities.SetWorkPermutation(workers, dto.FirstWorker, dto.WorkerCount)
	return task
}

// getNotificatorValues returns notifyTime, notifyInterval
func getNotificatorValues(task entities.Task) (interface{}, interface{}) {
	notificatorValues := reflect.Indirect(reflect.ValueOf(task.Notificator))
	notifyTime := notificatorValues.Field(0)
	notifyInterval := notificatorValues.Field(1)
	return notifyTime.Interface(), notifyInterval.Interface()
}

// getNotificatorValues returns  workers, workerCount, firstWorkerPtr
func getWokerPermutationValues(task entities.Task) (interface{}, interface{}, interface{}) {
	permutationValues := reflect.Indirect(reflect.ValueOf(task.Workers))
	workers := permutationValues.Field(0)
	workerCount := permutationValues.Field(1)
	firstWorkerPtr := permutationValues.Field(2)
	return workers.Interface(), workerCount.Interface(), firstWorkerPtr.Interface()
}
