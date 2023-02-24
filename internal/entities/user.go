package entities

import valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"

type User struct {
	ID          valueObjects.UserID
	ChannelID   valueObjects.ChannelID
	Credit      valueObjects.Money
	PhoneNumber valueObjects.PhoneNumber
	UserName    string
}
