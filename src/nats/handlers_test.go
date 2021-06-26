package nats

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type SpyCollection struct {

}

type mockDao interface {
	InsertAlarm() (*mongo.InsertOneResult, error)
}

type mockCollection interface {
	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}) (*mongo.SingleResult)
}

//type mockCollectionObj struct {}
//func (r mockCollectionObj) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
//	return
//}
//
//func TestAlarmStatusChangeHandler_InsertsAlarm(*testing.T) {
//
//
//	AlarmStatusChangeHandler(mockCollection)
//
//}
