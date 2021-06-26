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
	InsertAlarm(alarm AlarmStatusChangeEvent) (*mongo.InsertOneResult, error)
	GetActiveAlarms(userId string, from primitive.DateTime) ([]AlarmStatusChangeUpdate, error)
}

type AlarmDao struct {
	Collection *mongo.Collection
}

type AlarmStatusChangeEvent struct {
	AlarmID   string             `json:"alarmId" bson:"alarmId"`
	UserID    string             `json:"userId" bson:"userId"`
	Status    string             `json:"status" bson:"status"`
	ChangedAt primitive.DateTime `json:"changedAt" bson:"changedAt"`
}

type AlarmStatusChangeUpdate struct {
	AlarmID         string             `json:"alarmId" bson:"alarmId"`
	UserID          string             `json:"userId" bson:"userId"`
	Status          string             `json:"status" bson:"status"`
	LatestChangedAt primitive.DateTime `json:"latestChangedAt" bson:"changedAt"`
}

func (a AlarmDao) BuildAlarmIndexes() {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)

	indexModels := []mongo.IndexModel{
		{
			Keys: bson.M{"changedAt": 1},
			Options: nil,
		},
		{
			Keys: bson.M{"userId": 1},
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

func (a AlarmDao) InsertAlarm(alarm AlarmStatusChangeEvent) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//data, err := bson.Marshal(alarm)
	//if err != nil {
	//	return nil, err
	//}

	log.Printf("Inserting AlarmStatusChangeEvent record: %r", alarm)
	return a.Collection.InsertOne(ctx, alarm)
}

func (a AlarmDao) GetActiveAlarms(userId string, from primitive.DateTime) ([]AlarmStatusChangeUpdate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetLimit(50)
	findOptions.SetSort(bson.D{{ "changedAt", 1}})

	var results []AlarmStatusChangeUpdate

	filter := bson.D{
		{ "userId", userId },
		{"changedAt", bson.M{"$gt": from} },
		{"$or", []bson.M{
			bson.M{"status": "CRITICAL"},
			bson.M{"status": "ALARM"},
		}},
	}
	cur, err := a.Collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("GetActiveAlarms - lookup failed with Find error: %e", err)
		return nil, err
	}

	for cur.Next(ctx) {
		var alarmChange AlarmStatusChangeUpdate
		err := cur.Decode(&alarmChange)
		if err != nil {
			log.Printf("GetActiveAlarms - lookup failed with Decode error: %e", err)
			return nil, err
		}
		if alarmChange.LatestChangedAt >= from {
			results = append(results, alarmChange)
		}
	}

	if err := cur.Err(); err != nil {
		log.Printf("GetActiveAlarms - lookup failed with cursor error: %e", err)
	}

	return results, err
}