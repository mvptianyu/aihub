/*
@Project: aihub
@Module: core
@File : error.go
*/
package core

import "errors"

var (
	ErrChatCompletionInvalidModel       = errors.New("model is not supported")
	ErrChatCompletionStreamNotSupported = errors.New("streaming is not supported")
	ErrMessageContentFieldsMisused      = errors.New("message content fields are missing")
)
