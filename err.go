package ratelimiter

import "errors"

var (
	TimeIntervalErr = errors.New("time interval cannot less than 100ms")
	LimitErr        = errors.New("limit cannot less than 1")
)
