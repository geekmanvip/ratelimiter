package ratelimiter

import "errors"

var (
	TimeIntervalErr     = errors.New("time interval cannot less than 100ms")
	LimitErr            = errors.New("limit cannot less than 1")
	CapacityErr         = errors.New("capacity cannot less than 1")
	RateErr             = errors.New("rate cannot less than 1")
	CapacityLessRateErr = errors.New("capacity cannot less than rate")
)
