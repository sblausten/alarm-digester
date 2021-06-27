package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/sblausten/go-service/src/nats"
	"github.com/sblausten/go-service/src/dao"
	"github.com/sblausten/go-service/src/config"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Starting Digest Service...")
	config := config.BuildConfig()

	ctx, cancel := context.WithCancel(context.Background())

	dbClient := dao.BuildClient(config, ctx)
	defer dbClient.Disconnect(ctx)

	alarmsCollection := dao.GetCollection(dbClient, config.Db.Name, "alarms")
	digestCollection := dao.GetCollection(dbClient, config.Db.Name, "digest")
	digestDao := dao.DigestDao{Collection: digestCollection}
	alarmDao := dao.AlarmDao{Collection: alarmsCollection}

	digestDao.BuildDigestIndexes()
	alarmDao.BuildAlarmIndexes()

	natsSubscriber := nats.NatsSubscriber{Config: config, Context: ctx}

	go natsSubscriber.StartSubscriber(config.Nats.SubscriberSubjectAlarmStatusChange, nats.AlarmStatusChangeHandler(alarmDao))
	go natsSubscriber.StartSubscriber(config.Nats.SubscriberSubjectSendAlarmDigest, nats.SendAlarmDigestHandler(digestDao, alarmDao, config))

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

