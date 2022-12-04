package entities

import valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"

type User struct {
	ID          valueObjects.UserID
	ChannelID   valueObjects.ChannelID
	Credit      valueObjects.Money
	PhoneNumber valueObjects.PhoneNumber
	UserName    string
}
