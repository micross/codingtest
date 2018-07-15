package utils

import (
	"log"
	"time"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s\r", msg, err)
	}
}

func TomorrowUnix() int64 {
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	return t.AddDate(0, 0, 1).UnixNano() / 1000000
}
