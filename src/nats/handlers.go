package nats

import (
	"github.com/sblausten/go-service/src/config"
	"github.com/sblausten/go-service/src/dao"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

func AlarmStatusChangeHandler(alarmDao dao.AlarmDaoInterface) func(message dao.AlarmStatusChanged) {
	messageCounter := 0

	return func(message dao.AlarmStatusChanged) {
		messageCounter++
		log.Printf("[#%d] Received AlarmStatusChanged for [%s] with alarmId: '%s'", messageCounter, message.ChangedAt, message.AlarmID)

		_, err := alarmDao.InsertAlarm(message)
		if err != nil {
			log.Printf("Failed to save AlarmStatusChanged message with alarmId: %a for user: %u", message.AlarmID, message.UserID)
		} else {
			log.Printf("Saved AlarmStatusChanged message with alarmId: %a for user: %u", message.AlarmID, message.UserID)
		}
	}
}

func SendAlarmDigestHandler(digestDao dao.DigestDaoInterface, alarmDao dao.AlarmDaoInterface, config config.Config) func(message dao.SendAlarmDigest) {
	messageCounter := 0

	return func(message dao.SendAlarmDigest) {
		messageCounter++
		log.Printf("[#%d] Received SendAlarmDigest request with userId: '%u'", messageCounter, message.UserID)

		lastDigest, err := digestDao.GetLastDigest(message.UserID)
		getAlarmsFrom := lastDigest.RequestedAt
		if err != nil {
			minusOneMonth := time.Hour * -728
			getAlarmsFrom = primitive.NewDateTimeFromTime(time.Now().Add(minusOneMonth))
		}
		digestDao.InsertDigest(message)


		userAlarms, err := alarmDao.GetActiveAlarms(message.UserID, getAlarmsFrom)
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
