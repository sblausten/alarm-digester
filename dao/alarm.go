package dao

import (
	"context"
	"github.com/sblausten/go-service/models"
	"github.com/sblausten/go-service/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type AlarmDaoInterface interface {
	BuildAlarmIndexes()
	UpsertAlarm(alarm models.AlarmStatusChangeMessage) error
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
			Keys:    bson.M{"userId": 1},
			Options: nil,
		},
		{
			Keys:    bson.M{"alarmId": 1},
			Options: nil,
		},
		{
			Keys:    bson.M{"changedAt": 1},
			Options: nil,
		},
	}
	indexes, err := a.Collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Println("BuildAlarmIndexes - Error creating indexes:", err)
	} else {
		log.Printf("BuildAlarmIndexes - Created indexes %v on collection %s \n", indexes, a.Collection.Name())
	}
}

func (a AlarmDao) UpsertAlarm(alarmMessage models.AlarmStatusChangeMessage) (error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	alarmTime, err := util.GetNanoTimeFromString(alarmMessage.ChangedAt)
	if err != nil {
		log.Println("UpsertAlarm - Error parsing timestamp:", err)
		return err
	}

	unixAlarmTime := util.ToUnixNano(alarmTime)
	alarm := AlarmStatusChangeEvent{
		AlarmID:   alarmMessage.AlarmID,
		UserID:    alarmMessage.UserID,
		Status:    alarmMessage.Status,
		ChangedAt: unixAlarmTime,
	}
	filter := bson.D{
		{"userId", alarm.UserID},
		{"alarmId", alarm.AlarmID},
	}
	opts := options.Replace().SetUpsert(true)

	_, err = a.Collection.ReplaceOne(ctx, filter, alarm, opts)

	return err
}

func (a AlarmDao) GetActiveAlarms(userId string, from int64) ([]AlarmStatusChangeEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	timeFromFormatted := util.ConvertUnixToFormatted(from)
	log.Printf("GetActiveAlarms - Getting active alarms for %s from %s", userId, timeFromFormatted)

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"changedAt", 1}})

	var results []AlarmStatusChangeEvent

	filter := bson.D{
		{"userId", userId},
		{"changedAt", bson.M{"$gt": from}},
		{"$or", []bson.M{
			bson.M{"status": "CRITICAL"},
			bson.M{"status": "WARNING"},
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

