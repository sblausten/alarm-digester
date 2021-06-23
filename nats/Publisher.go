package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/sblausten/go-service/config"
	"log"
)

func PublishMessage(subject string, message AlarmDigest, config config.Config) {
	opts := []nats.Option{nats.Name("AlarmDigest Publisher")}

	nc, err := nats.Connect(config.Nats.ServerAddress, opts...)
	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	requestChanSend := make(chan *AlarmDigest)
	ec.BindSendChan(subject, requestChanSend)

	requestChanSend <- &message

	//nc.Publish(subject, message)
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Printf("PublishMessage - failed to publish message %m. Error was: '%e'\n", message, err)
	} else {
		log.Printf("PublishMessage - published message to subject %s", subject)
	}
}
