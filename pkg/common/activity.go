// Package common provides common functionality.
package common

import (
	"context"
	"time"
)

type ActivityContext struct {
	OperatingSystem string    `json:"operating_system"`
	Browser         string    `json:"browser"`
	IPAddress       string    `json:"ip_address"`
	UserAgent       string    `json:"user_agent"`
	Device          string    `json:"device"`
	CountryCode     string    `json:"country_code,omitempty"`
	Location        string    `json:"location"`
	Timestamp       time.Time `json:"timestamp"`
	RequestID       string    `json:"request_id"`
	Referer         string    `json:"referer"`
	Source          string    `json:"source,omitempty"`
	Locale          string    `json:"locale,omitempty"`
}

type ActivityContextKey struct{}

func WithActivity(ctx context.Context, activity *ActivityContext) context.Context {
	return context.WithValue(ctx, ActivityContextKey{}, activity)
}

func ActivityFromContext(ctx context.Context) ActivityContext {
	activity, ok := ctx.Value(ActivityContextKey{}).(*ActivityContext)
	if !ok {
		return ActivityContext{}
	}

	return *activity
}
