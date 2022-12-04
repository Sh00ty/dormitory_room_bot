package entities

import valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"

type Transaction struct {
	From   valueObjects.UserID
	To     valueObjects.UserID
	Amount valueObjects.Money
	ChatID valueObjects.ChannelID
}
