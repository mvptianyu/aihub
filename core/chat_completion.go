/*
@Project: aihub
@Module: core
@File : types.go
*/
package core

// TODO：参考：https://github.com/sashabaranov/go-openai.git

// CreateChatCompletionReq 参见https://platform.openai.com/docs/api-reference/chat/create
type CreateChatCompletionReq struct {
	Messages         []*Message `json:"messages"`
	Model            string     `json:"model"`
	FrequencyPenalty int        `json:"frequency_penalty,omitempty"`
	MaxTokens        int        `json:"max_tokens,omitempty"`
	PresencePenalty  int        `json:"presence_penalty,omitempty"`
	Stop             string     `json:"stop,omitempty"`
	Stream           bool       `json:"stream,omitempty"`
	Temperature      int        `json:"temperature,omitempty"`
	TopP             int        `json:"top_p,omitempty"`
	Tools            []*Tool    `json:"tools,omitempty"`
}

// CreateChatCompletionRsp 参见https://platform.openai.com/docs/api-reference/chat/create
type CreateChatCompletionRsp struct {
	Id      string                 `json:"id,omitempty"`
	Object  string                 `json:"object,omitempty"`
	Created int                    `json:"created,omitempty"`
	Model   string                 `json:"model,omitempty"`
	Choices []ChatCompletionChoice `json:"choices,omitempty"`
	Usage   struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		TotalTokens         int `json:"total_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
			AudioTokens  int `json:"audio_tokens"`
		} `json:"prompt_tokens_details"`
		CompletionTokensDetails struct {
			ReasoningTokens          int `json:"reasoning_tokens"`
			AudioTokens              int `json:"audio_tokens"`
			AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
			RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
		} `json:"completion_tokens_details"`
	} `json:"usage,omitempty"`
	ServiceTier       string `json:"service_tier,omitempty"`
	SystemFingerprint string `json:"system_fingerprint,omitempty"`
	Error             struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Param   interface{} `json:"param"`
		Code    interface{} `json:"code"`
	} `json:"error,omitempty"`
}

type CreateChatCompletionStream Stream[CreateChatCompletionRsp]

type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      *Message    `json:"message,omitempty"` // stream = false时返回
	Delta        *Message    `json:"delta,omitempty"`   // stream = true时返回
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}
