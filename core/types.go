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
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role        string        `json:"role"`
			Content     string        `json:"content"`
			Refusal     interface{}   `json:"refusal"`
			Annotations []interface{} `json:"annotations"`
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
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
	} `json:"usage"`
	ServiceTier       string `json:"service_tier"`
	SystemFingerprint string `json:"system_fingerprint"`
}

// ---------------------Message--------------------------
type MessageRoleType string

const (
	MessageRoleUser      MessageRoleType = "user"
	MessageRoleAssistant MessageRoleType = "assistant"
	MessageRoleSystem    MessageRoleType = "system"
	MessageRoleTool      MessageRoleType = "tool"
)

type MessageContentType string

const (
	MessageContentTypeText  MessageContentType = "text"
	MessageContentTypeImage MessageContentType = "image_url"
	MessageContentTypeAudio MessageContentType = "input_audio"
	MessageContentTypeFile  MessageContentType = "file"
)

type Message struct {
	Role       MessageRoleType   `json:"role"`
	Content    []*MessageContent `json:"content"`
	Name       string            `json:"name,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
}

type MessageContent struct {
	Type       MessageContentType   `json:"type"`
	Text       string               `json:"text,omitempty"`
	ImageUrl   *MessageContentImage `json:"image_url,omitempty"`
	InputAudio *MessageContentAudio `json:"input_audio,omitempty"`
	File       *MessageContentFile  `json:"file,omitempty"`
}

type MessageContentImage struct {
	URL string `json:"url"` // 图片url或base64数据
}
type MessageContentAudioFormat string

const (
	MessageContentAudioFormatMP3 MessageContentAudioFormat = "mp3"
	MessageContentAudioFormatWAV MessageContentAudioFormat = "wav"
)

type MessageContentAudio struct {
	Data   string                    `json:"data"`
	Format MessageContentAudioFormat `json:"format"` // mp3|wav
}

type MessageContentFile struct {
	FileData string `json:"file_data"` // base64数据
	FileName string `json:"format"`    // 文件名
}

// ---------------------------Tool------------------------------------
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunctionParametersType string

const (
	ToolFunctionParametersTypeText   ToolFunctionParametersType = "text"
	ToolFunctionParametersTypeObject ToolFunctionParametersType = "object"
)

type ToolFunction struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Parameters  *ToolFunctionParameters `json:"parameters,omitempty"`
	Strict      bool                    `json:"strict,omitempty"`
}

type ToolFunctionParameters struct {
	Type ToolFunctionParametersType `json:"type"`
}
