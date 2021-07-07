package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/sblausten/alarm-digester/app/config"
	"github.com/sblausten/alarm-digester/app/models"
	"log"
)

type PublisherInterface interface {
	PublishMessage(subject string, message models.AlarmDigest) error
}

type Publisher struct {
	Config config.Config
}

func (p Publisher) PublishMessage(subject string, message models.AlarmDigest) error {
	opts := []nats.Option{nats.Name("AlarmDigest Publisher")}

	nc, err := nats.Connect(p.Config.Nats.ServerAddress, opts...)
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
		log.Printf("PublishMessage - failed to publish message %v. Error was: '%v'\n", message, err)
	} else {
		log.Printf("PublishMessage - published digest to subject %s with %d alarms", subject, len(message.ActiveAlarms))
	}

	return err
}
