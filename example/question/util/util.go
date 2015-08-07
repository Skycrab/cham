package util

import (
	"time"
)

const (
	TIME_FORMATER = "2006-01-02 15:04:05"
)

func ParseTime(t string) time.Time {
	tt, err := time.Parse(TIME_FORMATER, t)
	if err != nil {
		tt = time.Unix(0, 0)
	}
	return tt
}

func FormatTime(t time.Time) string {
	return t.Format(TIME_FORMATER)
}
