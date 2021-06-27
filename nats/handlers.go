package nats

import (
	"github.com/sblausten/go-service/config"
	"github.com/sblausten/go-service/dao"
	"github.com/sblausten/go-service/models"
	"github.com/sblausten/go-service/util"
	"log"
	"time"
)

func AlarmStatusChangeHandler(alarmDao dao.AlarmDaoInterface) func(message models.AlarmStatusChangeMessage) {
	messageCounter := 0

	return func(message models.AlarmStatusChangeMessage) {
		messageCounter++
		log.Printf("AlarmStatusChangeHandler - [#%d] Received AlarmStatusChangeEvent for [%s] with alarmId: '%s'", messageCounter, message.ChangedAt, message.AlarmID)

		err := alarmDao.UpsertAlarm(message)
		if err != nil {
			log.Printf("AlarmStatusChangeHandler - Failed to save AlarmStatusChangeEvent message with alarmId: %s for user: %s. Error: %v", message.AlarmID, message.UserID, err)
		} else {
			log.Printf("AlarmStatusChangeHandler - Saved AlarmStatusChangeEvent message with alarmId: %s for user: %s", message.AlarmID, message.UserID)
		}
	}
}

func SendAlarmDigestHandler(
	digestDao dao.DigestDaoInterface,
	alarmDao dao.AlarmDaoInterface,
	publisher PublisherInterface,
	config config.Config) func(message dao.SendAlarmDigest) {

	messageCounter := 0
	return func(message dao.SendAlarmDigest) {
		messageCounter++
		log.Printf("[#%d] Received SendAlarmDigest request with userId: '%s'", messageCounter, message.UserId)

		lastDigest, err := digestDao.GetLastDigest(message.UserId)
		getAlarmsFrom := lastDigest.RequestedAt
		if err != nil {
			log.Printf("SendAlarmDigestHandler - Error finding last digest for user %s: %v", message.UserId, err)
			minusOneMonth := time.Hour * -728
			getAlarmsFrom = time.Now().UTC().Add(minusOneMonth).UnixNano()
			log.Printf("SendAlarmDigestHandler - Fetching alarms from one month ago %d", getAlarmsFrom)
		}
		err = digestDao.InsertDigest(message)
		if err != nil {
			log.Printf("SendAlarmDigestHandler - Error inserting digest request user %s: %v", message.UserId, err)
			return
		}

		userAlarms, err := alarmDao.GetActiveAlarms(message.UserId, getAlarmsFrom)
		if err != nil {
			log.Printf("SendAlarmDigestHandler - Error looking up Active alarms for user %s: %v", message.UserId, err)
			return
		}
		if len(userAlarms) == 0 {
			log.Printf("SendAlarmDigestHandler - Found no Active alarms from after %d for user %s", getAlarmsFrom, message.UserId)
		}

		activeAlarms := []models.ActiveAlarm{}

		for _, alarmChange := range userAlarms {
			time := util.ConvertUnixToFormatted(alarmChange.ChangedAt)

			publishedAlarm := models.ActiveAlarm{
				AlarmID:         alarmChange.AlarmID,
				Status:          alarmChange.Status,
				LatestChangedAt: time,
			}
			activeAlarms = append(activeAlarms, publishedAlarm)
		}

		alarmDigest := models.AlarmDigest{message.UserId, activeAlarms}

		publisher.PublishMessage(config.Nats.ProducerSubjectAlarmDigest, alarmDigest)
	}
}
