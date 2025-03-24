package sqs

import (
	"errors"
)

var ErrSizeLimit = errors.New("the size limit to publish is 10 messages")

type ErrSendMessage struct {
	Code    string
	Message string
}

var errSendMessage = &ErrSendMessage{}

func (e *ErrSendMessage) Error() string {
	return "status=%s message=%s"
}

func (e *ErrSendMessage) Is(target error) bool {
	return errors.Is(target, errSendMessage)
}
