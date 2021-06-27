package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/sblausten/go-service/src/config"
	"github.com/sblausten/go-service/src/models"
	"log"
)

func PublishMessage(subject string, message models.AlarmDigest, config config.Config) {
	opts := []nats.Option{nats.Name("AlarmDigest Publisher")}

	nc, err := nats.Connect(config.Nats.ServerAddress, opts...)
	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	requestChanSend := make(chan *models.AlarmDigest)
	ec.BindSendChan(subject, requestChanSend)

	requestChanSend <- &message
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Printf("PublishMessage - failed to publish message %m. Error was: '%e'\n", message, err)
	} else {
		log.Printf("PublishMessage - published digest to subject %s with %i alarms", subject, len(message.ActiveAlarms))
	}
}
