package ptr

import "time"

func Now() *time.Time {
	return New(time.Now())
}
