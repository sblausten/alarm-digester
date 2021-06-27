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
