package main

import "fmt"

type SendAlarmDigest struct {
	UserID string
}

type AlarmStatusChanged struct {
	AlarmID string
	UserID string
	Status string
	ChangedAt string
}

func main() {
	fmt.Println("Testing testing 123...")
}

