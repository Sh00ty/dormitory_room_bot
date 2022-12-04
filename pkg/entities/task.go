package entities

import (
	"time"

	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type TaskType int

const (
	// которые были с самого начала
	Default TaskType = iota

	// можно добавиться или отписаться от нее (тут количество исполнителей=все)
	Subscribed

	// одноразовая (после напоминания) умирает
	OneShot
)

type SubscribedStages int

const (
	// нельзя подписываться и отписываться(как обычная таска)
	Ready = iota

	// можно подписываться и отписываться
	Subscribing
)

// Task сущность задания/обязанности которые мы будем выполнять
type Task struct {
	ID valueObjects.TaskID

	// хранит и отвечает за тех кто должен исполнять задание(сама их мешает)
	Workers WorkerMaster

	// текстовое представление задачи
	Title string

	// хранит время нотификации
	Notificator Notificator

	Type TaskType

	Author string

	Sstage SubscribedStages
}

type WorkerMaster interface {
	GetWorkers() []valueObjects.UserID
	GenerateNext() error
	Unsub(worker valueObjects.UserID) error
	Sub(worker valueObjects.UserID) error
}

type Notificator interface {
	UpdateNotificationDate()
	GetNextNotifyTime() (time.Time, bool)
}

// CreateTask создает обязанность или задачу
// поля связанные со врменем опциональные
func CreateTask(
	id valueObjects.TaskID, title, author string,
	workers []valueObjects.UserID, workerCount uint32,
	notifyInterval *time.Duration, notifyTime *time.Time) Task {

	task := Task{
		ID:     id,
		Title:  title,
		Author: author,
	}

	task.Workers = CreateUserWorkPermutation(workers, workerCount)
	task.Notificator = CreateNotificator(notifyInterval, notifyTime)

	return task
}

func (t Task) IsOneShot() bool {
	return t.Type>>OneShot&1 == 1
}

func (t Task) IsSubscribed() bool {
	return t.Type>>Subscribed&1 == 1
}

// Notificate если время для оповещения наступило, то возвращает true и сам обновляется
func (t *Task) UpdateNotificationDate() {
	t.Notificator.UpdateNotificationDate()
}

// Getworkers получает перестановку исполлнителей для работы над задачей
func (t Task) Getworkers() ([]valueObjects.UserID, error) {
	Workers := t.Workers.GetWorkers()
	if len(Workers) == 0 {
		return nil, localerrors.ErrDoesntExist
	}

	return Workers, nil
}

func (t Task) GetNextNotifyTime() (time.Time, bool) {
	return t.Notificator.GetNextNotifyTime()
}

func (t *Task) MakeSubscribed() {
	t.Type = t.Type | (1 << Subscribed)
	t.Sstage = Subscribing
}

func (t *Task) MakeOneShot() {
	t.Notificator.UpdateNotificationDate()
	t.Type = t.Type | (1 << OneShot)
}

func (t *Task) EndSubs() {
	t.Sstage = Ready
}

func (t *Task) Unsub(woker valueObjects.UserID) error {
	return t.Workers.Unsub(woker)
}

func (t *Task) Sub(woker valueObjects.UserID) error {
	return t.Workers.Sub(woker)
}
