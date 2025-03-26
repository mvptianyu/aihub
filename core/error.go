/*
@Project: aihub
@Module: core
@File : error.go
*/
package core

import "errors"

var (
	ErrUnknown                          = errors.New("unknown error")
	ErrConfiguration                    = errors.New("invalid configuration")
	ErrChatCompletionInvalidModel       = errors.New("model is not supported")
	ErrChatCompletionStreamNotSupported = errors.New("streaming is not supported")
	ErrMessageContentFieldsMisused      = errors.New("message content fields are missing")
	ErrHTTPRequestURLInvalid            = errors.New("http request url invalid")
	ErrHTTPRequestBodyInvalid           = errors.New("http request body invalid")
	ErrHTTPResponseReadFailed           = errors.New("http response read err")
	ErrHTTPStatusCodeNon200Failed       = errors.New("http response statuscode not match 2xx err")
	ErrHTTPRequestTimeout               = errors.New("http request timeout")
)
