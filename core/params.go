/*
@Project: aihub
@Module: core
@File : params.go
*/
package core

// ChatCompletionReq 参见https://platform.openai.com/docs/api-reference/chat/create
type ChatCompletionReq struct {
	Messages         []ChatCompletionReqMessage `json:"messages"`
	Model            string                     `json:"model"`
	FrequencyPenalty int                        `json:"frequency_penalty,omitempty"`
	MaxTokens        int                        `json:"max_tokens,omitempty"`
	PresencePenalty  int                        `json:"presence_penalty,omitempty"`
	Stop             string                     `json:"stop,omitempty"`
	Stream           bool                       `json:"stream,omitempty"`
	Temperature      int                        `json:"temperature,omitempty"`
	TopP             int                        `json:"top_p,omitempty"`
	Tools            []ChatCompletionReqTool    `json:"tools,omitempty"`
}

type ChatCompletionReqMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
	Name    string `json:"name,omitempty"`
}

type ChatCompletionReqTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  struct {
			Type       string `json:"type"`
			Properties struct {
				Location struct {
					Type        string `json:"type"`
					Description string `json:"description"`
				} `json:"location"`
			} `json:"properties"`
			Required []string `json:"required"`
		} `json:"parameters"`
	} `json:"function"`
}
