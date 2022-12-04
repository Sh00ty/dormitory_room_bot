package localerrors

import (
	"errors"
	"fmt"

	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type localerr error

var (
	ErrDoesntExist localerr = errors.New("doesn't exists")

	ErrAlreadyExists localerr = errors.New("already exists")

	ErrDidntUpdated localerr = errors.New("can't update in db")

	ErrNothingToNotificate localerr = errors.New("there is nothing to update")

	ErrCantGetWorkers localerr = errors.New("can't get workers for task")

	ErrCantUpdateWorkers localerr = errors.New("can't update workers for task")

	ErrNotSyncedCache localerr = errors.New("cache not synced")

	ErrInvalidArgument localerr = errors.New("invalid cache argument")

	ErrTryAgain localerr = errors.New("try again")

	ErrNoWorkers localerr = errors.New("there is no workers for task")

	ErrNoItems localerr = errors.New("there is no items in this list")

	ErrNotSubTask localerr = errors.New("task isn't sub/unsub")

	ErrDidntSubscribed localerr = errors.New("you didn't subscribe to this task")
)

type UpdateCreditError struct {
	Err  error
	User valueObjects.UserID
}

func (e UpdateCreditError) Error() string {
	return fmt.Sprintf("%s : %s", e.Err.Error(), e.User)
}
