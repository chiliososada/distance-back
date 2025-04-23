package util

import (
	"errors"
	"fmt"
	"math"
	"time"
)

func SafeSlice[T any](s []T, from int, count int) ([]T, int, error) {
	if from >= len(s) {
		return nil, 0, errors.New("out of boundary")
	}

	end := int(math.Min(float64(from+count), float64(len(s))))
	return s[from:end], end, nil
}

func RoundToNextTokyoMidnight(t time.Time) time.Time {
	// Get the local timezone
	//loc := t.Location()
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		loc = time.UTC
	}
	fmt.Printf("loc: %v\n", loc)

	localT := t.In(loc)

	// Add 1 day to the current time
	nextDay := localT.AddDate(0, 0, 1)

	// Truncate to the beginning of that day (midnight)
	return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, loc).UTC()

}
