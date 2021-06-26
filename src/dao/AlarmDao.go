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

type AlarmDaoInterface interface {
	BuildAlarmIndexes()
	InsertAlarm(alarm AlarmStatusChanged) (*mongo.InsertOneResult, error)
	GetActiveAlarms(userId string, from primitive.DateTime) ([]AlarmStatusChanged, error)
}

type AlarmDao struct {
	Collection *mongo.Collection
}

type AlarmStatusChanged struct {
	_ID        primitive.ObjectID
	Alarm_ID   string
	User_ID    string
	Status     string
	Changed_At primitive.DateTime
}

func (a AlarmDao) BuildAlarmIndexes() {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)

	indexModels := []mongo.IndexModel{
		{
			Keys: bson.M{"changed_at": 1},
			Options: nil,
		},
		{
			Keys: bson.M{"user_id": 1},
			Options: nil,
		},
		{
			Keys: bson.M{"status": 1},
			Options: nil,
		},
	}
	indexes, err := a.Collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Println("Error creating indexs:", err)
	} else {
		fmt.Printf("Created indexes %i on collection %c \n", indexes, a.Collection.Name())
	}
}

func (a AlarmDao) InsertAlarm(alarm AlarmStatusChanged) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := bson.Marshal(alarm)
	if err != nil {
		return nil, err
	}

	log.Printf("Inserting AlarmStatusChanged record: %r", string(data))
	return a.Collection.InsertOne(ctx, data)
}

func (a AlarmDao) GetActiveAlarms(userId string, from primitive.DateTime) ([]AlarmStatusChanged, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetLimit(50)
	findOptions.SetSort(bson.D{{ "changed_at", 1}})

	var results []AlarmStatusChanged

	activeAlarmsForUser := bson.D{
		{ "user_id", userId },
		{"$or", []bson.M{
			bson.M{"status": "CRITICAL"},
			bson.M{"status": "ALARM"},
		}},
	}
	cur, err := a.Collection.Find(ctx, activeAlarmsForUser, findOptions)
	if err != nil {
		log.Printf("GetActiveAlarms - lookup failed with Find error: %e", err)
		return nil, err
	}

	for cur.Next(ctx) {
		var alarmChange AlarmStatusChanged
		err := cur.Decode(&alarmChange)
		if err != nil {
			log.Printf("GetActiveAlarms - lookup failed with Decode error: %e", err)
			return nil, err
		}
		if alarmChange.Changed_At >= from {
			results = append(results, alarmChange)
		}
	}

	if err := cur.Err(); err != nil {
		log.Printf("GetActiveAlarms - lookup failed with cursor error: %e", err)
	}

	return results, err
}