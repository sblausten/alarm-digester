package nats

import "github.com/sblausten/go-service/src/dao"

type AlarmDigest struct {
	UserID string
	ActiveAlarms []dao.AlarmStatusChangeUpdate
}

