package main

import (
	"context"
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/core"
)

func main() {
	ctx := context.Background()

	//

	// Create a new agent
	myAgent := aihub.NewAgent(&core.AgentConfig{
		Provider:     core.ProviderConfig{},
		SystemPrompt: "You are a helpful assistant.",
	})

	msg, err := myAgent.RunStream(
		ctx,
		"Why is the sky blue?",
	)

	for delta := range result.DeltaChan {
		print(delta)
	}
}
