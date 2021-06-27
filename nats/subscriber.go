package nats

import (
	"context"
	"fmt"
	"github.com/sblausten/go-service/config"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsSubscriberInterface interface {
	StartSubscriber(subject string, messageHandler nats.Handler)
}

type NatsSubscriber struct {
	Config config.Config
	Context context.Context
}

func (s NatsSubscriber) StartSubscriber(subject string, messageHandler nats.Handler) {

	subscriberName := fmt.Sprintf("%s Subscriber", subject)
	opts := []nats.Option{nats.Name(subscriberName)}
	opts = setupConnOptions(opts)

	nc, err := nats.Connect(s.Config.Nats.ServerAddress, opts...)
	if err != nil {
		log.Fatal(err)
	}
	encodedConnection, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}

	encodedConnection.QueueSubscribe(subject, s.Config.Nats.QueueGroup, messageHandler)
	encodedConnection.Flush()
	log.Printf( "Subscribed to Nats server for subject: %s \n", subject)

	if err := nc.LastError(); err != nil {
		nc.Drain()
		log.Fatal(err)
	}

	select {
	case <-s.Context.Done():
		if err := nc.Drain(); err != nil {
			log.Fatal(err)
		}
	}
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.DrainTimeout(20*time.Second))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectHandler(func(nc *nats.Conn) {
		log.Printf("Disconnected: will attempt reconnects for %.0fm", totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
		if s != nil {
			log.Printf("Async error in %q/%q: %v", s.Subject, s.Queue, err)
		} else {
			log.Printf("Async error outside subscription: %v", err)
		}
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}