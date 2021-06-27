package dao

import (
	"context"
	"github.com/sblausten/go-service/src/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type AlarmDaoInterface interface {
	BuildAlarmIndexes()
	InsertAlarm(alarm models.AlarmStatusChangeMessage) (*mongo.InsertOneResult, error)
	GetActiveAlarms(userId string, from int64) ([]AlarmStatusChangeEvent, error)
}

type AlarmDao struct {
	Collection *mongo.Collection
}

type AlarmStatusChangeEvent struct {
	AlarmID   string `bson:"alarmId"`
	UserID    string `bson:"userId"`
	Status    string `bson:"status"`
	ChangedAt int64  `bson:"changedAt"`
}

func (a AlarmDao) BuildAlarmIndexes() {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"changedAt": 1},
			Options: nil,
		},
		{
			Keys:    bson.M{"userId": 1},
			Options: nil,
		},
		{
			Keys:    bson.M{"status": 1},
			Options: nil,
		},
	}
	indexes, err := a.Collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Println("Error creating indexs:", err)
	} else {
		log.Printf("Created indexes %i on collection %c \n", indexes, a.Collection.Name())
	}
}

func (a AlarmDao) InsertAlarm(alarmMessage models.AlarmStatusChangeMessage) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	time, err := time.Parse(time.RFC3339Nano, alarmMessage.ChangedAt)
	if err != nil {
		log.Println("InsertAlarm - Error parsing timestamp:", err)
		return nil, err
	}

	unixTime := time.Unix()
	alarm := AlarmStatusChangeEvent{
		AlarmID:   alarmMessage.AlarmID,
		UserID:    alarmMessage.UserID,
		Status:    alarmMessage.Status,
		ChangedAt: unixTime,
	}

	log.Printf("Inserting AlarmStatusChangeEvent record: %r", alarm)
	return a.Collection.InsertOne(ctx, alarm)
}

func (a AlarmDao) GetActiveAlarms(userId string, from int64) ([]AlarmStatusChangeEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Printf("Getting active alarms for %u from %t", userId, time.Unix(0, from).UTC().Format(time.RFC3339))

	findOptions := options.Find()
	findOptions.SetLimit(50)
	findOptions.SetSort(bson.D{{"changedAt", 1}})

	var results []AlarmStatusChangeEvent

	filter := bson.D{
		{"userId", userId},
		//{"changedAt", bson.M{"$gt": from}},
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
		var alarm AlarmStatusChangeEvent
		err := cur.Decode(&alarm)
		if err != nil {
			log.Printf("GetActiveAlarms - lookup failed with Decode error: %e", err)
			return nil, err
		}

		results = append(results, alarm)
	}

	if err := cur.Err(); err != nil {
		log.Printf("GetActiveAlarms - lookup failed with cursor error: %e", err)
	}

	return results, err
}
