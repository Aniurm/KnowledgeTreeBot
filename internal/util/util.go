package util

import "time"

func ParseTimestamp(timestamp float64) (int, int) {
	t := time.Unix(int64(timestamp/1000), 0)
	return t.Year(), int(t.Month())
}
