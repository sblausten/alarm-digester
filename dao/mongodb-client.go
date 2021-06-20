package dao

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func buildClient(env string) *mongo.Client {
	/*
	   Connect to my cluster
	*/

	dbUrl := getConnectionUrl(env)

	client, err := mongo.NewClient(options.Client().ApplyURI(dbUrl))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
}


func getConnectionUrl(env string) string {
	if env == "dev" {
		return "localhost:27017"
	} else {
		return "mongodb+srv://netdata-test:<password>@cluster0.1un1e.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	}
}