package entities

import valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"

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
