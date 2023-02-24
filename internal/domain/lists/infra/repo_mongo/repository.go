package lists

import (
	"context"
	"errors"
	"time"

	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	bsonIds "gopkg.in/mgo.v2/bson"
)

type repository struct {
	c *mongo.Collection
}

func NewListsRepository(ctx context.Context, addr, dbName, password, user string) (*repository, func(ctx context.Context) error, error) {

	// credential := options.Credential{
	// 	Username: user,
	// 	Password: password,
	// }

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI(addr).SetMinPoolSize(minPoolSize).SetAuth(credential))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(addr))
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel = context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		err2 := client.Disconnect(ctx)
		if err2 != nil {
			logger.Errorf("can't disconnect from mongo : %v", err2)
		}
		return nil, nil, err
	}

	collection := client.Database(dbName).Collection("lists")

	return &repository{collection}, client.Disconnect, nil
}

func (r repository) GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.List, error) {
	cur, err := r.c.Find(ctx, bson.M{"channel_id": channelID})
	if err != nil {
		return nil, err
	}
	if cur.RemainingBatchLength() == 0 {
		return nil, localerrors.ErrDoesntExist
	}

	lists := make([]listDTO, 0)

	err = cur.All(ctx, &lists)
	if err != nil {
		return nil, err
	}

	return toEntLists(lists), nil
}

// list

func (r repository) GetList(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID) (entities.List, error) {
	res := r.c.FindOne(ctx, bson.M{"channel_id": channelID, "id": listID})

	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return entities.List{}, localerrors.ErrDoesntExist
		}
		return entities.List{}, res.Err()
	}
	var list listDTO
	err := res.Decode(&list)
	if err != nil {
		return entities.List{}, err
	}
	return entities.List{
		ID:        valueObjects.ListID(list.ID),
		ChannelID: valueObjects.ChannelID(list.ChannelID),
		Items:     toEntItems(list.Items),
	}, nil
}

func (r repository) CreateList(ctx context.Context, channelID valueObjects.ChannelID, list entities.List) error {
	res := r.c.FindOne(ctx, bson.M{"channel_id": channelID, "id": list.ID})

	if res.Err() != nil && !errors.Is(res.Err(), mongo.ErrNoDocuments) {
		if !errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return localerrors.ErrAlreadyExists
		}
		return res.Err()
	}
	_, err := r.c.InsertOne(ctx, listDTO{
		MID:       bsonIds.NewObjectId(),
		ID:        string(list.ID),
		ChannelID: int64(list.ChannelID),
		Items:     toDbItems(list.Items),
	})
	return err
}

func (r repository) DeleteList(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID) error {

	res, err := r.c.DeleteOne(ctx, bson.M{"channel_id": channelID, "id": listID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return localerrors.ErrDoesntExist
	}
	return nil
}

// items

func (r repository) DeleteItem(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, index uint) error {
	res := r.c.FindOne(ctx, bson.M{"channel_id": channelID, "id": listID})

	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return localerrors.ErrDoesntExist
		}
		return res.Err()
	}
	var list listDTO
	err := res.Decode(&list)
	if err != nil {
		return err
	}
	if index >= uint(len(list.Items)) {
		return localerrors.ErrDoesntExist
	}

	list.Items = append(list.Items[:index], list.Items[index+1:]...)

	res2, err := r.c.ReplaceOne(ctx, bson.M{"channel_id": channelID, "id": listID}, list)
	if err != nil {
		return err
	}
	if res2.ModifiedCount == 0 {
		return localerrors.ErrDidntUpdated
	}
	return nil
}

func (r repository) AddItem(ctx context.Context, channelID valueObjects.ChannelID, listID valueObjects.ListID, item entities.Item) error {
	res := r.c.FindOne(ctx, bson.M{"channel_id": channelID, "id": listID})

	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return localerrors.ErrDoesntExist
		}
		return res.Err()
	}
	var list listDTO
	err := res.Decode(&list)
	if err != nil {
		return err
	}
	list.Items = append(list.Items, itemDTO{
		Author: item.Author,
		Name:   item.Name,
	})

	res2, err := r.c.ReplaceOne(ctx, bson.M{"channel_id": channelID, "id": listID}, list)
	if err != nil {
		return err
	}
	if res2.ModifiedCount == 0 {
		return localerrors.ErrDidntUpdated
	}
	return nil
}
