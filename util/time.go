package util

import "time"

func GetCurrentUTCTimeAsUnixNano() int64 {
	return time.Now().UTC().UnixNano()
}

func GetNanoTimeFromString(timeString string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, timeString)
}

func ToUnixNano(utcAlarmTime time.Time) int64 {
	return utcAlarmTime.UTC().UnixNano()
}

func ConvertUnixToFormatted(from int64) string {
	return time.Unix(0, from).UTC().Format(time.RFC3339Nano)
}