package sqs

import (
	"sync"

	"github.com/aawadallak/go-core-kit/core/broker"
)

type attributes struct {
	sync.RWMutex
	attributes map[string][]string
}

var _ broker.Attributes = (*attributes)(nil)

func newAttributes() *attributes {
	return &attributes{
		attributes: make(map[string][]string),
	}
}

func (a *attributes) Add(key, value string) {
	a.attributes[key] = append(a.attributes[key], value)
}

func (a *attributes) Get(key string) string {
	values, ok := a.attributes[key]
	if !ok || len(values) == 0 {
		return ""
	}
	return values[0]
}

func (a *attributes) Lookup(key string) (string, bool) {
	values, ok := a.attributes[key]
	if !ok || len(values) == 0 {
		return "", false
	}
	return values[0], true
}

func (a *attributes) Delete(key string) {
	delete(a.attributes, key)
}

func (a *attributes) Values() map[string][]string {
	return a.attributes
}
