package entities

import valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"

type Credit struct {
	ChannelID valueObjects.ChannelID `db:"channel_id"`
	UserID    valueObjects.UserID    `db:"user_id"`
	Credit    valueObjects.Money     `db:"credit"`
}

func NewDebit(channelID valueObjects.ChannelID, userID valueObjects.UserID, credit valueObjects.Money) *Credit {
	return &Credit{
		ChannelID: channelID,
		UserID:    userID,
		Credit:    credit,
	}
}
