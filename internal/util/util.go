package util

import (
	"errors"
	"math"
)

func SafeSlice[T any](s []T, from int, count int) ([]T, int, error) {
	if from >= len(s) {
		return nil, 0, errors.New("out of boundary")
	}

	end := int(math.Min(float64(from+count), float64(len(s))))
	return s[from:end], end, nil
}
