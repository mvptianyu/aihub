package aihub

import "errors"

var (
	ErrUnknown                     = errors.New("unknown error")
	ErrConfiguration               = errors.New("invalid agent or provider configuration")
	ErrToolRegisterRepeat          = errors.New("tool function name register repeated")
	ErrProviderRateLimit           = errors.New("provider trigger rate limit")
	ErrAgentRunTimeout             = errors.New("agent run timeout")
	ErrCallNameNotMatch            = errors.New("not found matched call name with mcp/tool entry")
	ErrMCPResponseEmpty            = errors.New("mcp call response empty")
	ErrChatCompletionOverMaxStep   = errors.New("chat request over max step quit")
	ErrMessageContentFieldsMisused = errors.New("message content fields are missing")
	ErrHTTPRequestURLInvalid       = errors.New("http request url invalid")
	ErrHTTPRequestBodyInvalid      = errors.New("http request body invalid")
	ErrHTTPRequestTimeout          = errors.New("http request timeout")
)
