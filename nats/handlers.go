package nats

import (
	"context"
	"github.com/sblausten/go-service/config"
	"github.com/sblausten/go-service/dao"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func AlarmStatusChangeHandler(alarmsCollection *mongo.Collection, cancellable context.Context) func(message dao.AlarmStatusChanged) {
	messageCounter := 0

	return func(message dao.AlarmStatusChanged) {
		messageCounter++
		log.Printf("[#%d] Received AlarmStatusChanged for [%s] with alarmId: '%s'", messageCounter, message.ChangedAt, message.AlarmID)

		_, err := dao.InsertAlarm(alarmsCollection, message, cancellable)
		if err != nil {
			log.Printf("Failed to save AlarmStatusChanged message with alarmId: %a for user: %u", message.AlarmID, message.UserID)
		} else {
			log.Printf("Saved AlarmStatusChanged message with alarmId: %a for user: %u", message.AlarmID, message.UserID)
		}

		return
	}
}

func SendAlarmDigestHandler(digestCollection *mongo.Collection, alarmsCollection *mongo.Collection, cancellable context.Context, config config.Config) func(message dao.SendAlarmDigest) {
	messageCounter := 0

	return func(message dao.SendAlarmDigest) {
		messageCounter++
		log.Printf("[#%d] Received SendAlarmDigest request with userId: '%u'", messageCounter, message.UserID)

		lastDigest, err := dao.GetLastDigest(digestCollection, message.UserID, cancellable)
		getAlarmsFrom := lastDigest.RequestedAt
		if err != nil {
			minusOneMonth := time.Hour * -728
			getAlarmsFrom = primitive.NewDateTimeFromTime(time.Now().Add(minusOneMonth))
		}
		dao.InsertDigest(digestCollection, message, cancellable)


		userAlarms, err := dao.GetAlarms(alarmsCollection, message.UserID, getAlarmsFrom)
		if err != nil {
			return
		}


		activeAlarms := []ActiveAlarm{}

		for _, alarmChange := range userAlarms {
			activeAlarm := ActiveAlarm{alarmChange.AlarmID, alarmChange.Status, alarmChange.ChangedAt.Time().UTC().String()}
			activeAlarms = append(activeAlarms, activeAlarm)
		}

		alarmDigest := AlarmDigest{message.UserID, activeAlarms}

		PublishMessage(config.Nats.AlarmDigestSubject, alarmDigest, config)
	}
}
