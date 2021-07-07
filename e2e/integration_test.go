package e2e

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

type AlarmDigest struct {
	UserID       string
	ActiveAlarms []ActiveAlarm
}

type ActiveAlarm struct {
	AlarmID         string
	Status          string
	LatestChangedAt string
}

type AlarmStatusChangeMessage struct {
	AlarmID   string `json:"alarmId"`
	UserID    string `json:"userId"`
	Status    string `json:"status"`
	ChangedAt string `json:"changedAt"`
}

type SendAlarmDigest struct {
	UserId      string `json:"userId"`
	RequestedAt int64 `json:"requestedAt"`
}

var receiverSubject = "AlarmDigest"
var alarmSubject = "AlarmStatusChanged"
var digestSubject = "SendAlarmDigest"

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func TestAlarmDigest_ForUser(t *testing.T) {
	t.Skip()
	// Var defs
	assert := assert.New(t)
	now := time.Now().UTC()
	first := formatTime(now)
	second := formatTime(now.Add(10 * time.Second))
	requestedAt := now.Add(20 * time.Second).UnixNano()
	testUserId := randUserId()

	nc, ec := setupNats(t)
	defer nc.Close()

	// Input
	alarm1 := AlarmStatusChangeMessage{AlarmID: "1", UserID: testUserId, Status: "CRITICAL", ChangedAt: first}
	alarm2 := AlarmStatusChangeMessage{AlarmID: "2", UserID: testUserId, Status: "CRITICAL", ChangedAt: second}
	alarm3 := AlarmStatusChangeMessage{AlarmID: "3", UserID: "2", Status: "CRITICAL", ChangedAt: formatTime(now)}
	digestRequest := SendAlarmDigest{UserId: testUserId, RequestedAt: requestedAt}

	// Expected
	event1 := ActiveAlarm{AlarmID: "1", Status: "CRITICAL", LatestChangedAt: first}
	event2 := ActiveAlarm{AlarmID: "2", Status: "CRITICAL", LatestChangedAt: second}
	activeAlarms := []ActiveAlarm{event2, event1}
	expected := AlarmDigest{UserID: testUserId, ActiveAlarms: activeAlarms,}
	expectedAsJson, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("TestAlarmDigest - failed to json encode expected %v", err)
	}

	// Run Test
	sub, err := nc.SubscribeSync(receiverSubject)
	if err != nil {
		t.Fatalf("TestAlarmDigest - Could not subscribe to subject %s", receiverSubject)
	}

	ec.Publish(alarmSubject, alarm1)
	ec.Publish(alarmSubject, alarm2)
	ec.Publish(alarmSubject, alarm3)

	ec.Publish(digestSubject, digestRequest)

	msg, err := sub.NextMsg(5 * time.Second)
	if err != nil {
		t.Fatalf("TestAlarmDigest - Timed out waiting for message on subject %s: %v", receiverSubject, err)
	}

	// Assert
	assert.Equal(string(expectedAsJson), string(msg.Data), "should return ordered active alarms of user requested")
}

func TestAlarmDigest_FromLastDigest(t *testing.T) {
	// Var defs
	assert := assert.New(t)
	now := time.Now().UTC()
	first := formatTime(now)
	second := formatTime(now.Add(20 * time.Second))
	requestedAt1 := now.Add(10 * time.Second).UnixNano()
	requestedAt2 := now.Add(30 * time.Second).UnixNano()
	testUserId := randUserId()

	nc, ec := setupNats(t)
	defer nc.Close()

	// Input
	alarm1 := AlarmStatusChangeMessage{AlarmID: "1", UserID: testUserId, Status: "CRITICAL", ChangedAt: first}
	alarm2 := AlarmStatusChangeMessage{AlarmID: "2", UserID: testUserId, Status: "CRITICAL", ChangedAt: second}
	digestRequest1 := SendAlarmDigest{UserId: testUserId, RequestedAt: requestedAt1}
	digestRequest2 := SendAlarmDigest{UserId: testUserId, RequestedAt: requestedAt2}

	// Expected
	event1 := ActiveAlarm{AlarmID: "1", Status: "CRITICAL", LatestChangedAt: first}
	event2 := ActiveAlarm{AlarmID: "2", Status: "CRITICAL", LatestChangedAt: second}

	expected1 := AlarmDigest{UserID: testUserId, ActiveAlarms: []ActiveAlarm{event1}}
	expected2 := AlarmDigest{UserID: testUserId, ActiveAlarms: []ActiveAlarm{event2}}

	// Run Test
	sub, err := nc.SubscribeSync(receiverSubject)
	if err != nil {
		t.Fatalf("TestAlarmDigest - Could not subscribe to subject %s", receiverSubject)
	}

	ec.Publish(alarmSubject, alarm1)
	time.Sleep(1 * time.Second)
	ec.Publish(digestSubject, digestRequest1)

	msg, err := sub.NextMsg(5 * time.Second)
	if err != nil {
		t.Fatalf("TestAlarmDigest - Timed out waiting for message on subject %s: %v", receiverSubject, err)
	}

	expectedJson1 := toJsonString(expected1, t)
	assert.Equal(expectedJson1, string(msg.Data), "should return active alarm for user")

	ec.Publish(alarmSubject, alarm2)
	time.Sleep(1 * time.Second)
	ec.Publish(digestSubject, digestRequest2)

	msg, err = sub.NextMsg(5 * time.Second)
	if err != nil {
		t.Fatalf("TestAlarmDigest - Timed out waiting for message on subject %s: %v", receiverSubject, err)
	}
	expectedJson2 := toJsonString(expected2, t)
	assert.Equal(expectedJson2, string(msg.Data), "should return active alarms created since last digest")
}

func TestAlarmDigest_NoActiveAlarms(t *testing.T) {
	t.Skip()
	// Var defs
	assert := assert.New(t)
	now := time.Now().UTC()
	first := formatTime(now)
	second := formatTime(now.Add(10 * time.Second))
	requestedAt := now.Add(20 * time.Second).UnixNano()
	testUserId := randUserId()

	nc, ec := setupNats(t)
	defer nc.Close()

	// Input
	alarm1 := AlarmStatusChangeMessage{AlarmID: "1", UserID: testUserId, Status: "CRITICAL", ChangedAt: first}
	alarm2 := AlarmStatusChangeMessage{AlarmID: "1", UserID: testUserId, Status: "CLEARED", ChangedAt: second}
	digestRequest := SendAlarmDigest{UserId: testUserId, RequestedAt: requestedAt}

	// Expected
	expected := AlarmDigest{UserID: testUserId, ActiveAlarms: []ActiveAlarm{}}
	expectedAsJson, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("TestAlarmDigest - failed to json encode expected %v", err)
	}

	// Run Test
	sub, err := nc.SubscribeSync(receiverSubject)
	if err != nil {
		t.Fatalf("TestAlarmDigest - Could not subscribe to subject %s", receiverSubject)
	}

	ec.Publish(alarmSubject, alarm1)
	ec.Publish(alarmSubject, alarm2)

	ec.Publish(digestSubject, digestRequest)

	msg, err := sub.NextMsg(5 * time.Second)
	if err != nil {
		t.Fatalf("TestAlarmDigest - Timed out waiting for message on subject %s: %v", receiverSubject, err)
	}

	// Assert
	assert.Equal(string(expectedAsJson), string(msg.Data), "should update alarm to cleared and not return in digest")
}

func randUserId() string {
	return strconv.Itoa(rand.Intn(100))
}

func toJsonString(digest AlarmDigest, t *testing.T) string {
	expectedAsJson, err := json.Marshal(digest)
	if err != nil {
		t.Fatalf("TestAlarmDigest - failed to json encode expected %v", err)
	}
	return string(expectedAsJson)
}

func setupNats(t *testing.T) (*nats.Conn, *nats.EncodedConn) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("TestAlarmDigest - Connecting to %s failed: %v", nats.DefaultURL, err)
	}
	ec, _ := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		t.Fatalf("TestAlarmDigest - Encoded connection to %s failed: %v", nats.DefaultURL, err)
	}
	nc.Flush()
	return nc, ec
}
