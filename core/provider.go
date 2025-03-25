package core

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub/providers/openai/client"
	"github.com/mvptianyu/aihub/types"
	"log/slog"
)

// LLM提供商
type provider struct {
	cfg *ProviderConfig

	// client is the internal Ollama HTTP client
	client *client.OpenAIClient
}

func (p *provider) CreateChatCompletion(ctx context.Context, request *CreateChatCompletionReq) (response *CreateChatCompletionRsp, err error) {
	// TODO implement me
	panic("implement me")
}

func (p *provider) CreateChatCompletionStream(ctx context.Context, request *CreateChatCompletionReq) (stream *CreateChatCompletionStream, err error) {
	// TODO implement me
	panic("implement me")
}

func NewProvider(cfg *ProviderConfig) IProvider {
	ins := &provider{
		cfg: cfg,
	}
	return ins
}

// Generate implements the LLMProvider interface for basic responses
func (p *provider) Generate(ctx context.Context, opts *types.GenerateOptions) (*types.Message, error) {
	slog.Info("Generate request received", "modelID", p.model.ID)

	resp, err := p.client.Chat(ctx, &client.ChatRequest{
		Model:    p.model.ID,
		Messages: opts.Messages,
		Tools:    opts.Tools,
	})

	if err != nil {
		p.logger.Error(err.Error(), "Error calling client chat method", err)
		return nil, fmt.Errorf("error calling client chat method: %w", err)
	}

	return &types.Message{
		Role:      types.AssistantMessageRole,
		Content:   resp.Message.Content,
		ToolCalls: resp.Message.ToolCalls,
	}, nil
}

// GenerateStream streams the response token by token
func (p *provider) GenerateStream(ctx context.Context, opts *types.GenerateOptions) (<-chan *types.Message, <-chan string, <-chan error) {
	p.logger.Info("Starting stream generation", "modelID", p.model.ID)

	return p.client.ChatStream(ctx, &client.ChatRequest{
		Model:    p.model.ID,
		Messages: opts.Messages,
		Tools:    opts.Tools,
	})
}
