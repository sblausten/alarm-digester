package dao

import (
	"context"
	"github.com/sblausten/go-service/src/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"strings"
)

func BuildClient(config config.Config, ctx context.Context) *mongo.Client {
	dbUrl := getConnectionUrl(config)

	client, err := mongo.NewClient(options.Client().ApplyURI(dbUrl))
	if err != nil {
		log.Fatalf("Could not build MongoDb client: %e", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Could not connect to MongoDb database: %e", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Could not ping MongoDb database: %e", err)
	}

	return client
}

func GetCollection(client *mongo.Client, databaseName string, collectionName string) *mongo.Collection {
	collection := client.Database(databaseName).Collection(collectionName)
	return collection
}

func getConnectionUrl(config config.Config) string {
	if config.Env == "dev" {
		return config.Db.LocalAddress
	} else {
		if config.Db.Password == "" {
			log.Fatalf("No password found. Cannot connect to remote db in env: %e", config.Env)
		}
		return strings.Replace(config.Db.RemoteAddress, "<password>", config.Db.Password, 0)
	}
}