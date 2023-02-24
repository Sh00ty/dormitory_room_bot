package entities

import valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"

type Transaction struct {
	From   valueObjects.UserID
	To     valueObjects.UserID
	Amount valueObjects.Money
	ChatID valueObjects.ChannelID
}
