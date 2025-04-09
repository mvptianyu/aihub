/*
@Project: aihub
@Module: core
@File : error.go
*/
package aihub

import "errors"

var (
	ErrUnknown                          = errors.New("unknown error")
	ErrConfiguration                    = errors.New("invalid agent or provider configuration")
	ErrToolRegisterEmpty                = errors.New("tool function name not find")
	ErrToolRegisterRepeat               = errors.New("tool function name register repeated")
	ErrProviderRateLimit                = errors.New("provider trigger rate limit")
	ErrAgentRunTimeout                  = errors.New("agent run timeout")
	ErrCallNameNotMatch                 = errors.New("not found matched call name with mcp/tool entry")
	ErrMCPResponseEmpty                 = errors.New("mcp call response empty")
	ErrChatCompletionInvalidModel       = errors.New("model is not supported")
	ErrChatCompletionStreamNotSupported = errors.New("streaming is not supported")
	ErrChatCompletionOverMaxStep        = errors.New("chat request over max step quit")
	ErrMessageContentFieldsMisused      = errors.New("message content fields are missing")
	ErrHTTPRequestURLInvalid            = errors.New("http request url invalid")
	ErrHTTPRequestBodyInvalid           = errors.New("http request body invalid")
	ErrHTTPResponseReadFailed           = errors.New("http response read err")
	ErrHTTPStatusCodeNon200Failed       = errors.New("http response statuscode not match 2xx err")
	ErrHTTPRequestTimeout               = errors.New("http request timeout")
)
