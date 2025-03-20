package main

import (
	"context"
	"github.com/mvptianyu/aihub/pkg/agent"
	"github.com/mvptianyu/aihub/providers/openai"
	"github.com/mvptianyu/aihub/providers/openai/models"
)

func main() {
	ctx := context.Background()

	// Create an openai provider
	provider := openai.NewProvider(&openai.ProviderOpts{
		Logger: logger,
	})
	provider.UseModel(ctx, models.GPT4_O)

	// Create a new agent
	myAgent := agent.NewAgent(&agent.NewAgentConfig{
		Provider:     provider,
		Logger:       logger,
		SystemPrompt: "You are a helpful assistant.",
	})

	result := myAgent.RunStream(
		ctx,
		agent.WithInput("Why is the sky blue?"),
	)

	for delta := range result.DeltaChan {
		print(delta)
	}
}
