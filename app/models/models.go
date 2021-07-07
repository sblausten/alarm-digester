package models

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
	UserId      string `json:"userId" bson:"userId"`
	RequestedAt int64 `json:"requestedAt" bson:"requestedAt"`
}
