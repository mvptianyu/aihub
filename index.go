/*
@Project: aihub
@Module: aihub
@File : index.go
*/
package aihub

import (
	"github.com/mvptianyu/aihub/core"
	inmemory "github.com/mvptianyu/aihub/pkg/memory/inmem"
	"github.com/mvptianyu/aihub/providers/openai/client"
	"github.com/mvptianyu/aihub/types"
	"log/slog"
)

// NewAgent creates a new agent with the given provider
func NewAgent(cfg *core.AgentConfig) core.IAgent {
	if cfg.MaxSteps == 0 {
		// set a sane default max steps
		cfg.MaxSteps = 20
	}

	if cfg.MemoryStorer == nil {
		cfg.MemoryStorer = inmemory.NewInMemoryMemStore()
	}

	return &Agent{
		provider: config.Provider,
		tools:    make(map[string]types.Tool),
		memory:   config.Memory,
		maxSteps: config.MaxSteps,
		logger:   config.Logger,
	}
}

// NewProvider creates a new Ollama provider
func NewProvider(cfg *core.ProviderConfig) core.IProvider {
	slog.Debug("Creating new OpenAI provider", slog.Any("cfg", cfg))

	return core.NewProvider(cfg)

	client := client.NewClient(
	// TODO - need to enable local env variable, not just through opt
	// client.WithAPIKey(opts.APIKey),
	)

	return &Provider{
		client: client,
		logger: *opts.Logger,
	}
}
