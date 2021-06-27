package nats

import (
	"github.com/golang/mock/gomock"
	"github.com/sblausten/go-service/mocks"
	"github.com/sblausten/go-service/config"
	"github.com/sblausten/go-service/dao"
	"testing"
	"time"
)

//var requestedAt int64 = time.Now().UnixNano()
//
//type MockDigestDao struct {
//	Collection *mongo.Collection
//}
//type MockAlarmDao struct {
//	Collection *mongo.Collection
//}
//type MockPublisher struct {
//	Config config.Config
//}
//
//func (d MockDigestDao) BuildDigestIndexes() {}
//func (d MockDigestDao) InsertDigest(digest dao.SendAlarmDigest) (error) {
//	return nil
//}
//func (d MockDigestDao) GetLastDigest(userId string) (dao.SendAlarmDigest, error) {
//
//	digest := dao.SendAlarmDigest{UserId: userId, RequestedAt: requestedAt}
//	return digest, nil
//}
//
//func (d MockAlarmDao) BuildAlarmIndexes() {}
//func (d MockAlarmDao) UpsertAlarm(alarm models.AlarmStatusChangeMessage) error {
//	return nil
//}
//func (d MockAlarmDao) GetActiveAlarms(userId string, from int64) ([]dao.AlarmStatusChangeEvent, error) {
//	event1 := dao.AlarmStatusChangeEvent{AlarmID: "1", UserID: "1", Status: "CRITICAL", ChangedAt: requestedAt}
//	event2 := event1
//	event2.AlarmID = "2"
//	events := []dao.AlarmStatusChangeEvent{event1, event2}
//
//	return events, nil
//}
//
//func (p MockPublisher) PublishMessage(subject string, message models.AlarmDigest)  {}
//

//func TestAlarmStatusChangeHandler_InsertsAlarm(*testing.T) {
//
//	handler := AlarmStatusChangeHandler(MockAlarmDao{})
//
//}

func TestSendAlarmDigestHandler_InsertsAlarm(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := config.Config{Nats: config.NatsConfig{ProducerSubjectAlarmDigest: "Send" }}
	message := dao.SendAlarmDigest{UserId: "1", RequestedAt: time.Now().UnixNano()}
	mockPublisher := mocks.NewMockPublisherInterface(ctrl)
	mockDigestDao := mocks.NewMockDigestDaoInterface(ctrl)
	mockAlarmDao := mocks.NewMockAlarmDaoInterface(ctrl)

	handler := SendAlarmDigestHandler(mockDigestDao, mockAlarmDao, mockPublisher, config)

	handler(message)

	mockPublisher.EXPECT().
		PublishMessage(config.Nats.ProducerSubjectAlarmDigest, message)
}
