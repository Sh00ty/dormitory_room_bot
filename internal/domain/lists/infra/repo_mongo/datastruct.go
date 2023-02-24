package lists

import (
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
	"gopkg.in/mgo.v2/bson"
)

type listDTO struct {
	MID       bson.ObjectId `bson:"_id"`
	ID        string        `bson:"id"`
	ChannelID int64         `bson:"channel_id"`
	Items     []itemDTO     `bson:"items"`
}

type itemDTO struct {
	Author      string `bson:"author"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
}

func toEntItems(items []itemDTO) (res []entities.Item) {
	res = make([]entities.Item, 0, len(items))
	for _, item := range items {
		res = append(res, entities.Item{
			Author: item.Author,
			Name:   item.Name,
		})
	}
	return
}

func toDbItems(items []entities.Item) (res []itemDTO) {
	res = make([]itemDTO, 0, len(items))
	for _, item := range items {
		res = append(res, itemDTO{
			Author: item.Author,
			Name:   item.Name,
		})
	}
	return
}

func toEntLists(lists []listDTO) (res []entities.List) {
	res = make([]entities.List, 0, len(lists))

	for _, list := range lists {
		res = append(res, entities.List{
			ID:        valueObjects.ListID(list.ID),
			ChannelID: valueObjects.ChannelID(list.ChannelID),
			Items:     toEntItems(list.Items),
		})
	}
	return
}
