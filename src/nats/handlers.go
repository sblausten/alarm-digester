package nats

import (
	"context"
	"github.com/sblausten/go-service/src/config"
	"github.com/sblausten/go-service/src/dao"
	"github.com/sblausten/go-service/src/models"
	"log"
	"time"
)

func AlarmStatusChangeHandler(alarmDao dao.AlarmDaoInterface) func(message models.AlarmStatusChangeMessage) {
	messageCounter := 0

	return func(message models.AlarmStatusChangeMessage) {
		messageCounter++
		log.Printf("AlarmStatusChangeHandler - [#%d] Received AlarmStatusChangeEvent for [%s] with alarmId: '%s'", messageCounter, message.ChangedAt, message.AlarmID)

		_, err := alarmDao.InsertAlarm(message)
		if err != nil {
			log.Printf("AlarmStatusChangeHandler - Failed to save AlarmStatusChangeEvent message with alarmId: %a for user: %u. Error: ", message.AlarmID, message.UserID, err)
		} else {
			log.Printf("AlarmStatusChangeHandler - Saved AlarmStatusChangeEvent message with alarmId: %a for user: %u", message.AlarmID, message.UserID)
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
			log.Printf("SendAlarmDigestHandler - Error finding last digest for user %u: %e", message.UserId, err)
			minusOneMonth := time.Hour * -728
			getAlarmsFrom = time.Now().UTC().Add(minusOneMonth).UnixNano()
			log.Printf("SendAlarmDigestHandler - Fetching alarms from one month ago %d", getAlarmsFrom)
		}
		_, err = digestDao.InsertDigest(message)
		if err != nil {
			log.Printf("SendAlarmDigestHandler - Error inserting digest request user %u: %e", message.UserId, err)
			return
		}

		userAlarms, err := alarmDao.GetActiveAlarms(message.UserId, getAlarmsFrom)
		if err != nil {
			log.Printf("SendAlarmDigestHandler - Error looking up Active alarms for user %u: %e", message.UserId, err)
			return
		}
		if len(userAlarms) == 0 {
			log.Printf("SendAlarmDigestHandler - Found no Active alarms from after %d for user %u", getAlarmsFrom, message.UserId)
		}

		activeAlarms := []models.ActiveAlarm{}

		for _, alarmChange := range userAlarms {
			formattedTime1 := time.Unix(0, alarmChange.ChangedAt).UTC().Format(time.RFC3339Nano)

			//formattedTime2, err := time.Parse(time.RFC3339Nano, alarmChange.ChangedAt)
			//if err != nil {
			//	log.Printf("GetActiveAlarms - Parsing alarm %i timestamp from db failed: %e", alarmChange.AlarmID, err)
			//	return
			//}

			publishedAlarm := models.ActiveAlarm{
				AlarmID: alarmChange.AlarmID,
				Status: alarmChange.Status,
				LatestChangedAt: formattedTime1,
			}
			activeAlarms = append(activeAlarms, publishedAlarm)
		}

		alarmDigest := models.AlarmDigest{message.UserId, activeAlarms}

		PublishMessage(config.Nats.ProducerSubjectAlarmDigest, alarmDigest, config)
	}
}
