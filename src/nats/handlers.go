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
		log.Printf("[#%d] Received AlarmStatusChanged for [%s] with alarmId: '%s'", messageCounter, message.Changed_At, message.Alarm_ID)

		_, err := alarmDao.InsertAlarm(message)
		if err != nil {
			log.Printf("Failed to save AlarmStatusChanged message with alarmId: %a for user: %u", message.Alarm_ID, message.User_ID)
		} else {
			log.Printf("Saved AlarmStatusChanged message with alarmId: %a for user: %u", message.Alarm_ID, message.User_ID)
		}
	}
}

func SendAlarmDigestHandler(digestDao dao.DigestDaoInterface, alarmDao dao.AlarmDaoInterface, config config.Config) func(message dao.SendAlarmDigest) {
	messageCounter := 0

	return func(message dao.SendAlarmDigest) {
		messageCounter++
		log.Printf("[#%d] Received SendAlarmDigest request with userId: '%u'", messageCounter, message.User_ID)

		lastDigest, err := digestDao.GetLastDigest(message.User_ID)
		getAlarmsFrom := lastDigest.Requested_At
		if err != nil {
			minusOneMonth := time.Hour * -728
			getAlarmsFrom = primitive.NewDateTimeFromTime(time.Now().Add(minusOneMonth))
		}
		digestDao.InsertDigest(message)


		userAlarms, err := alarmDao.GetActiveAlarms(message.User_ID, getAlarmsFrom)
		if err != nil {
			return
		}


		activeAlarms := []ActiveAlarm{}

		for _, alarmChange := range userAlarms {
			activeAlarm := ActiveAlarm{alarmChange.Alarm_ID, alarmChange.Status, alarmChange.Changed_At.Time().UTC().String()}
			activeAlarms = append(activeAlarms, activeAlarm)
		}

		alarmDigest := AlarmDigest{message.User_ID, activeAlarms}

		PublishMessage(config.Nats.AlarmDigestSubject, alarmDigest, config)
	}
}
