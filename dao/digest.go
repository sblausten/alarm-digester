package dao

import (
	"context"
	"github.com/sblausten/go-service/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type DigestDaoInterface interface {
	BuildDigestIndexes()
	InsertDigest(digest SendAlarmDigest) error
	GetLastDigest(userId string) (SendAlarmDigest, error)
}

type DigestDao struct {
	Collection *mongo.Collection
}

type SendAlarmDigest struct {
	UserId      string `json:"userId" bson:"userId"`
	RequestedAt int64 `json:"requestedAt" bson:"requestedAt"`
}

func (d DigestDao) BuildDigestIndexes() {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)

	indexModels := []mongo.IndexModel{
		{
			Keys: bson.M{"requestedAt": 1},
			Options: nil,
		},
		{
			Keys: bson.M{"userId": 1},
			Options: nil,
		},
	}

	indexes, err := d.Collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Println("BuildDigestIndexes - Error creating indexes:", err)
	} else {
		log.Printf("BuildDigestIndexes - Created indexes %i on collection %c \n", indexes, d.Collection.Name())
	}
}

func (d DigestDao) InsertDigest(digest SendAlarmDigest) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	digest.RequestedAt = util.GetCurrentUTCTimeAsUnixNano()

	data, err := bson.Marshal(digest)
	if err != nil {
		return err
	}

	log.Printf("InsertDigest - saving digest request for user: %u", digest.UserId)
	res, err := d.Collection.InsertOne(ctx, data)
	if err != nil {
		log.Printf("InsertDigest - insert failed with error: %e", err)
	} else {
		log.Printf("InsertDigest - successfully inserted Digest %i \n", res)
	}

	return err
}

func (d DigestDao) GetLastDigest(userId string) (SendAlarmDigest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lastDigest SendAlarmDigest
	filter := bson.D{{"userId", userId}}
	findOptions := options.FindOne()
	findOptions.SetSort(bson.D{{"requestedAt", 1}})

	err := d.Collection.FindOne(ctx, filter, findOptions).Decode(&lastDigest)
	if err != nil {
		log.Printf("GetLastDigest - lookup failed with error: %e", err)
	} else {
		log.Printf("GetLastDigest - Found previous Digest: %+v\n", lastDigest)
	}

	return lastDigest, err
}