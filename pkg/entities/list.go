package entities

import valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"

type Item struct {
	Author string
	Name   string
}

type List struct {
	ID        valueObjects.ListID
	ChannelID valueObjects.ChannelID
	Items     []Item
}
