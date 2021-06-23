package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/sblausten/go-service/nats"
	"github.com/sblausten/go-service/dao"
	"github.com/sblausten/go-service/config"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Starting Digest Service...")
	config := config.BuildConfig()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	dbClient := dao.BuildClient(config, ctx)
	defer dbClient.Disconnect(ctx)

	alarmsCollection := dao.GetCollection(dbClient, config.Db.Name, "alarms")
	digestCollection := dao.GetCollection(dbClient, config.Db.Name, "digest")
	dao.BuildAlarmIndexes(alarmsCollection)
	dao.BuildDigestIndexes(digestCollection)

	go nats.StartNatsSubscriber(
			"AlarmStatusChanged",
			config,
			nats.AlarmStatusChangeHandler(alarmsCollection, ctx))

	go nats.StartNatsSubscriber(
			"SendAlarmDigest",
			config,
			nats.SendAlarmDigestHandler(digestCollection, alarmsCollection, ctx, config))

	var (
		shutdown    = make(chan os.Signal, 1)
		serverError = make(chan error, 1)
	)

	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
		case <-shutdown:
			cancel()
			log.Println("Terminate signal received")
		case err := <-serverError:
			cancel()
			log.Printf("Server error, unable to start: %v", err)
	}
}

