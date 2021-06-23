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

type SendAlarmDigest struct {
	_ID primitive.ObjectID
	UserID string
	RequestedAt primitive.DateTime
}

func BuildDigestIndexes(collection *mongo.Collection) {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)

	indexModels := []mongo.IndexModel{
		{
			Keys: bson.M{"requestedat": 1},
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

func InsertDigest(collection *mongo.Collection, digest SendAlarmDigest, ctx context.Context) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	digest.RequestedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

	data, err := bson.Marshal(digest)
	if err != nil {
		return nil, err
	}

	log.Printf("InsertDigest - inserting record: %r", string(data))
	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		log.Printf("InsertDigest - insert failed with error: %e", err)
	} else {
		log.Printf("InsertDigest - successfully inserted Digest %i \n", res.InsertedID)
	}

	return res, err
}

func GetLastDigest(collection *mongo.Collection, userId string, ctx context.Context) (SendAlarmDigest, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var lastDigest SendAlarmDigest
	filter := bson.D{{"userid", userId}}
	findOptions := options.FindOne()
	findOptions.SetSort(bson.D{{"requestedat", 1}})

	err := collection.FindOne(ctx, filter, findOptions).Decode(&lastDigest)
	if err != nil {
		log.Printf("GetLastDigest - lookup failed with error: %e", err)
	} else {
		log.Printf("GetLastDigest - Found previous Digest: %+v\n", lastDigest)
	}

	return lastDigest, err
}