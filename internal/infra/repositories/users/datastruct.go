package user

import (
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type userDTO struct {
	ID          valueObjects.UserID      `db:"id"`
	PhoneNumber valueObjects.PhoneNumber `db:"phone_number"`
	Username    string                   `db:"username"`
}

func UserListFromDTOList(userDTO []userDTO) (res []entities.User) {
	res = make([]entities.User, 0, len(userDTO))

	for _, u := range userDTO {
		res = append(res, entities.User{
			ID:          valueObjects.UserID(u.ID),
			PhoneNumber: valueObjects.PhoneNumber(u.PhoneNumber),
			UserName:    u.Username,
		})
	}
	return
}
