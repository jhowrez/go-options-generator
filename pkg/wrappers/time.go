package wrappers

import (
	"log"
	"time"
)

func MustParseDuration(durStr string) time.Duration {
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		log.Panicf("string '%s' is not a valid duration", durStr)
	}
	return dur
}
