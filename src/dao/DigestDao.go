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

type DigestDaoInterface interface {
	BuildDigestIndexes()
	InsertDigest(digest SendAlarmDigest) (*mongo.InsertOneResult, error)
	GetLastDigest(userId string) (SendAlarmDigest, error)
}

type DigestDao struct {
	Collection *mongo.Collection
}

type SendAlarmDigest struct {
	_ID          primitive.ObjectID
	User_ID      string
	Requested_At primitive.DateTime
}

func (d DigestDao) BuildDigestIndexes() {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)

	indexModels := []mongo.IndexModel{
		{
			Keys: bson.M{"requested_at": 1},
			Options: nil,
		},
		{
			Keys: bson.M{"user_id": 1},
			Options: nil,
		},
	}
	indexes, err := d.Collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Println("Error creating indexs:", err)
	} else {
		fmt.Printf("Created indexes %i on collection %c \n", indexes, d.Collection.Name())
	}
}

func (d DigestDao) InsertDigest(digest SendAlarmDigest) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	digest.Requested_At = primitive.NewDateTimeFromTime(time.Now().UTC())

	data, err := bson.Marshal(digest)
	if err != nil {
		return nil, err
	}

	log.Printf("InsertDigest - inserting record: %r", string(data))
	res, err := d.Collection.InsertOne(ctx, data)
	if err != nil {
		log.Printf("InsertDigest - insert failed with error: %e", err)
	} else {
		log.Printf("InsertDigest - successfully inserted Digest %i \n", res.InsertedID)
	}

	return res, err
}

func (d DigestDao) GetLastDigest(userId string) (SendAlarmDigest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lastDigest SendAlarmDigest
	filter := bson.D{{"user_id", userId}}
	findOptions := options.FindOne()
	findOptions.SetSort(bson.D{{"requested_at", 1}})

	err := d.Collection.FindOne(ctx, filter, findOptions).Decode(&lastDigest)
	if err != nil {
		log.Printf("GetLastDigest - lookup failed with error: %e", err)
	} else {
		log.Printf("GetLastDigest - Found previous Digest: %+v\n", lastDigest)
	}

	return lastDigest, err
}