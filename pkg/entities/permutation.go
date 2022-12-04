package entities

import (
	mrand "math/rand"
	"time"

	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

// userWorkPermutation перестановка для распределения того кто должен работать
type userWorkPermutation struct {
	Users          []valueObjects.UserID
	WorkerCount    uint32
	FirstWorkerPtr uint32
}

// Create генерирует рандомную перестановку заданного размера из заданных юзеров
func CreateUserWorkPermutation(Users []valueObjects.UserID, count uint32) *userWorkPermutation {
	var u userWorkPermutation
	u.Users = Users
	ul := uint32(len(Users))
	if ul == 0 {
		return &u
	}
	if ul < count {
		u.WorkerCount = ul
		return &u
	}
	mrand.Seed(time.Now().UnixNano())
	u.FirstWorkerPtr = mrand.Uint32() % ul
	if count != 0 {
		u.WorkerCount = count
	} else {
		u.WorkerCount = ul
	}
	return &u
}

// Set создает прям такую перестановку
func SetWorkPermutation(Users []valueObjects.UserID, firstWorker, count uint32) *userWorkPermutation {
	var u userWorkPermutation
	u.Users = make([]valueObjects.UserID, 0, len(Users))
	u.Users = append(u.Users, Users...)
	u.FirstWorkerPtr = firstWorker
	u.WorkerCount = count
	return &u
}

func min(a uint32, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// GetWorkers возвращает тех кто сейчас выполняет задание
func (u userWorkPermutation) GetWorkers() (workers []valueObjects.UserID) {
	workersLen := min(u.WorkerCount, uint32(len(u.Users)))
	workers = make([]valueObjects.UserID, 0, workersLen)

	for i := uint32(0); i < workersLen; i++ {
		workerInd := (i + u.FirstWorkerPtr) % uint32(len(u.Users))
		workers = append(workers, u.Users[workerInd])
	}

	return workers
}

// GenerateNext очев
func (u *userWorkPermutation) GenerateNext() error {
	if len(u.Users) == 0 {
		return localerrors.ErrNoWorkers
	}
	u.FirstWorkerPtr = (u.FirstWorkerPtr + 1) % uint32(len(u.Users))
	return nil
}

func (u *userWorkPermutation) Unsub(worker valueObjects.UserID) error {
	if len(u.Users) == 0 {
		return localerrors.ErrNoWorkers
	}
	for i, user := range u.Users {
		if user == worker {

			u.Users = append(u.Users[:i], u.Users[i+1:]...)
			u.WorkerCount--
			return nil

		}
	}
	return localerrors.ErrDidntSubscribed
}

func (u *userWorkPermutation) Sub(worker valueObjects.UserID) error {
	for _, user := range u.Users {
		if user == worker {
			return localerrors.ErrAlreadyExists
		}
	}
	u.WorkerCount++
	u.Users = append(u.Users, worker)
	return nil
}
