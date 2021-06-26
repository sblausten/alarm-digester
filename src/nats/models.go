package nats

type AlarmStatusChanged struct {
	AlarmID string
	UserID string
	Status string
	ChangedAt string
}

type SendAlarmDigest struct {
	UserID string
}

type ActiveAlarm struct {
	AlarmID       string
	Status        string
	LastChangedAt string
}

type AlarmDigest struct {
	UserID string
	ActiveAlarms []ActiveAlarm
}

