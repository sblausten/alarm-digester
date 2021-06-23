package dao

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type AlarmStatusChanged struct {
	_ID primitive.ObjectID
	AlarmID string
	UserID string
	Status string
	ChangedAt primitive.DateTime
}

func BuildAlarmIndexes(collection *mongo.Collection) {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)

	indexModels := []mongo.IndexModel{
		{
			Keys: bson.M{"changedat": 1},
			Options: nil,
		},
		{
			Keys: bson.M{"userid": 1},
			Options: nil,
		},
	}
	indexes, err := collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Println("Error creating indexs:", err)
	} else {
		fmt.Printf("Created indexes %i on collection %c \n", indexes, collection.Name())
	}
}

func InsertAlarm(collection *mongo.Collection, alarm AlarmStatusChanged, ctx context.Context) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	data, err := bson.Marshal(alarm)
	if err != nil {
		return nil, err
	}

	log.Printf("Inserting AlarmStatusChanged record: %r", string(data))
	return collection.InsertOne(ctx, data)
}

func GetAlarms(collection *mongo.Collection, userId string, from primitive.DateTime) ([]AlarmStatusChanged, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetLimit(50)
	findOptions.SetSort(bson.D{{ "changedat", 1}})

	var results []AlarmStatusChanged

	filter := bson.D{{ "userid", userId }}
	cur, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("GetAlarms - lookup failed with Find error: %e", err)
		return nil, err
	}

	for cur.Next(ctx) {
		var alarmChange AlarmStatusChanged
		err := cur.Decode(&alarmChange)
		if err != nil {
			log.Printf("GetAlarms - lookup failed with Decode error: %e", err)
			return nil, err
		}
		if alarmChange.ChangedAt >= from {
			results = append(results, alarmChange)
		}
	}

	if err := cur.Err(); err != nil {
		log.Printf("GetAlarms - lookup failed with cursor error: %e", err)
	}

	return results, err
}