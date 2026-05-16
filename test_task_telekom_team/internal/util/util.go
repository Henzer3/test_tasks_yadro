package util

import (
	"fmt"
	"time"
)

func ParseTimeDuration(s string) (time.Duration, error) {
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		return 0, err
	}

	duration := time.Duration(t.Hour())*time.Hour +
		time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second

	return duration, nil
}

func FormatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
