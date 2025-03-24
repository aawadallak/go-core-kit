package cache

import "time"

type Item struct {
	Key       string
	Value     any
	ExpiresIn time.Duration
}
