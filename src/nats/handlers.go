package nats

import (
	"context"
	"github.com/sblausten/go-service/src/config"
	"github.com/sblausten/go-service/src/dao"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

func AlarmStatusChangeHandler(alarmDao dao.AlarmDaoInterface) func(message dao.AlarmStatusChangeEvent) {
	messageCounter := 0

	return func(message dao.AlarmStatusChangeEvent) {
		messageCounter++
		log.Printf("[#%d] Received AlarmStatusChangeEvent for [%s] with alarmId: '%s'", messageCounter, message.ChangedAt, message.AlarmID)

		_, err := alarmDao.InsertAlarm(message)
		if err != nil {
			log.Printf("Failed to save AlarmStatusChangeEvent message with alarmId: %a for user: %u", message.AlarmID, message.UserID)
		} else {
			log.Printf("Saved AlarmStatusChangeEvent message with alarmId: %a for user: %u", message.AlarmID, message.UserID)
		}
	}
}

func SendAlarmDigestHandler(digestDao dao.DigestDaoInterface, alarmDao dao.AlarmDaoInterface, config config.Config) func(message dao.SendAlarmDigest) {
	messageCounter := 0
	return func(message dao.SendAlarmDigest) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		messageCounter++
		log.Printf("[#%d] Received SendAlarmDigest request with userId: '%u'", messageCounter, message.UserId)

		lastDigest, err := digestDao.GetLastDigest(message.UserId)
		getAlarmsFrom := lastDigest.RequestedAt
		if err != nil {
			minusOneMonth := time.Hour * -728
			getAlarmsFrom = primitive.NewDateTimeFromTime(time.Now().Add(minusOneMonth))
		}
		digestDao.InsertDigest(message)


		userAlarms, err := alarmDao.GetActiveAlarms(message.UserId, getAlarmsFrom)
		if err != nil {
			return
		}


		activeAlarms := []dao.AlarmStatusChangeUpdate{}

		for _, alarmChange := range userAlarms {
			activeAlarms = append(activeAlarms, alarmChange)
		}

		alarmDigest := AlarmDigest{message.UserId, activeAlarms}

		PublishMessage(config.Nats.ProducerSubjectAlarmDigest, alarmDigest, config)
		return
	}
}
