package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/sblausten/go-service/src/config"
	"github.com/sblausten/go-service/src/dao"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

var testTime = time.Date(2021, 6, 1,1,1,1,1, time.UTC)
var testConfig = config.Config{}

func PublishMsg(subject string, body []byte) {
	nc, _ := nats.Connect(nats.DefaultURL)
	c, _ := nats.NewEncodedConn(nc, "json")
	defer c.Close()

	me := &dao.AlarmStatusChanged{
		AlarmID: "2345",
		UserID: "1234",
		Status: "Critical",
		ChangedAt: primitive.NewDateTimeFromTime(testTime),
	}
	c.Publish(subject, me)

}





func TestStartNatsSubscriber(*testing.T) {
	testSubject := "Test"
	body := []byte("{\"message\":\"testing\"")



	StartNatsSubscriber(testSubject, )
	PublishMsg(testSubject, body)


}
