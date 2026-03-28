package idem

import (
	"fmt"
	"strings"
)

type keyConfig struct {
	action        string
	entityID      string
	correlationID string
}

type KeyOption func(*keyConfig)

func WithAction(value string) KeyOption {
	return func(cfg *keyConfig) {
		cfg.action = value
	}
}

func WithEntityID(value string) KeyOption {
	return func(cfg *keyConfig) {
		cfg.entityID = value
	}
}

func WithCorrelationID(value string) KeyOption {
	return func(cfg *keyConfig) {
		cfg.correlationID = value
	}
}

func BuildOperationKey(opts ...KeyOption) string {
	cfg := &keyConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	return fmt.Sprintf("%s:%s:%s",
		normalizeKeySegment(cfg.action, "unknown_action"),
		normalizeKeySegment(cfg.entityID, "unknown_entity"),
		normalizeKeySegment(cfg.correlationID, "unknown_correlation"),
	)
}

func normalizeKeySegment(value, fallback string) string {
	v := strings.TrimSpace(strings.ToLower(value))
	if v == "" {
		return fallback
	}

	replacer := strings.NewReplacer(
		":", "_",
		" ", "_",
		"\n", "_",
		"\t", "_",
	)
	v = replacer.Replace(v)
	if v == "" {
		return fallback
	}

	return v
}
