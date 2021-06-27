package nats

import (
	"github.com/golang/mock/gomock"
	"github.com/sblausten/go-service/mocks"
	"github.com/sblausten/go-service/config"
	"github.com/sblausten/go-service/dao"
	"github.com/sblausten/go-service/models"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"
)

var requestedAt int64 = time.Now().UnixNano()
var formatedTime string = time.Unix(0, requestedAt).UTC().Format(time.RFC3339Nano)

//
type MockDigestDao struct {
	Collection *mongo.Collection
}
type MockAlarmDao struct {
	Collection *mongo.Collection
}

func (d MockDigestDao) BuildDigestIndexes() {}
func (d MockDigestDao) InsertDigest(digest dao.SendAlarmDigest) error {
	return nil
}
func (d MockDigestDao) GetLastDigest(userId string) (dao.SendAlarmDigest, error) {
	digest := dao.SendAlarmDigest{UserId: userId, RequestedAt: requestedAt}
	return digest, nil
}

func (d MockAlarmDao) BuildAlarmIndexes() {}
func (d MockAlarmDao) UpsertAlarm(alarm models.AlarmStatusChangeMessage) error {
	return nil
}
func (d MockAlarmDao) GetActiveAlarms(userId string, from int64) ([]dao.AlarmStatusChangeEvent, error) {
	events := getTwoAlarms()

	return events, nil
}

func getTwoAlarms() []dao.AlarmStatusChangeEvent {
	event1 := dao.AlarmStatusChangeEvent{AlarmID: "1", UserID: "1", Status: "CRITICAL", ChangedAt: requestedAt}
	event2 := event1
	event2.AlarmID = "2"
	events := []dao.AlarmStatusChangeEvent{event1, event2}
	return events
}

func getTwoMessageAlarms() []models.ActiveAlarm {
	event1 := models.ActiveAlarm{AlarmID: "1", Status: "CRITICAL", LatestChangedAt: formatedTime}
	event2 := event1
	event2.AlarmID = "2"
	events := []models.ActiveAlarm{event1, event2}
	return events
}

func TestSendAlarmDigestHandler_PublishesAlarms(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userId := "1"

	config := config.Config{Nats: config.NatsConfig{ProducerSubjectAlarmDigest: "Send"}}
	message := dao.SendAlarmDigest{UserId: userId, RequestedAt: requestedAt}

	activeAlarms := getTwoMessageAlarms()
	expected := models.AlarmDigest{
		UserID:       userId,
		ActiveAlarms: activeAlarms,
	}
	mockPublisher := mocks.NewMockPublisherInterface(ctrl)
	mockPublisher.EXPECT().
		PublishMessage(config.Nats.ProducerSubjectAlarmDigest, expected)

	handler := SendAlarmDigestHandler(MockDigestDao{}, MockAlarmDao{}, mockPublisher, config)

	handler(message)
}

func TestAlarmStatusChangeHandler_SavesAlarm(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expected := models.AlarmStatusChangeMessage{
		UserID:    "1",
		AlarmID:   "2",
		Status:    "WARNING",
		ChangedAt: formatedTime,
	}
	mockAlarmDao := mocks.NewMockAlarmDaoInterface(ctrl)
	mockAlarmDao.EXPECT().
		UpsertAlarm(expected)

	handler := AlarmStatusChangeHandler(mockAlarmDao)

	handler(expected)
}
