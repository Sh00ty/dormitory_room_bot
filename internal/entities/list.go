package entities

import valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"

type Item struct {
	Author string
	Name   string
}

type List struct {
	ID        valueObjects.ListID
	ChannelID valueObjects.ChannelID
	Items     []Item
}
