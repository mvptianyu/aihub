/*
@Project: aihub
@Module: core
@File : interface.go
*/
package core

import (
	"context"
	"github.com/mvptianyu/aihub/types"
)

// IProvider LLM模型提供商相关能力
type IProvider interface {
	// GetCapabilities returns what features this provider supports through a
	// core.Capabilities struct. A provider may return an error if it cannot
	// construct or query for its capabilities.
	GetCapabilities(ctx context.Context) (*types.Capabilities, error)

	// UseModel takes a context and a model string ID (i.e., "qwen2.5") and configuration
	// options through a core.ModelKnobs struct. It returns:
	//
	// 1. An "ok" boolean defining if the provider supports the given model by
	//    the given options.
	// 2. The constructed core.Model itself.
	// 3. An error.
	//
	// A provider implementation may choose to return (true, nil, error) where
	// some pre-check, pre-authentication, or query to the API failed causing an
	// error despite the Model itself being supported.
	UseModel(ctx context.Context, model *types.Model) error

	// Generate uses the provider to generate a new message given the core.GenerateOptions
	Generate(ctx context.Context, opts *types.GenerateOptions) (*types.Message, error)

	// Generate uses the provider to stream messages. It returns:
	// * a *types.Message channel which should have complete messages to be consumed by providers.
	// * a string channel which are the streaming deltas from the provider.
	// * an error channel to surface any errors during streaming execution.j
	GenerateStream(ctx context.Context, opts *types.GenerateOptions) (<-chan *types.Message, <-chan string, <-chan error)
}

// IAgent interface defines the core capabilities required for an agent
type IAgent interface {
	// Run executes the agent's main loop with the given input until a stop condition is met
	Run(ctx context.Context, input string, stopCondition types.AgentStopCondition) ([]*types.AgentRunAggregator, error)

	// RunStream supports a streaming channel from a provider
	RunStream(ctx context.Context, input string, stopCondition types.AgentStopCondition) (<-chan types.AgentRunAggregator, <-chan string, <-chan error)

	// Step executes a single step of the agent's logic based on a given role
	Step(ctx context.Context, message types.Message) (*types.Message, error)

	// SendMessages provides a simpler interface for chat-style interactions
	SendMessages(ctx context.Context, content string) (*types.Message, error)

	// AddTool adds a new tool to the agent's capabilities
	AddTool(tool *types.Tool) error

	// GetTools returns the current set of available tools
	GetTools() []*types.Tool

	// Middleware functionality

	// ---------------------

	// RegisterMiddleware adds a middleware to the processing chain
	RegisterMiddleware(middle *Middleware) error

	// RemoveMiddleware removes a middleware by name
	RemoveMiddleware(name string) error

	// GetMiddleware returns a middleware by name
	GetMiddleware(name string) (Middleware, bool)

	// ListMiddleware returns all registered middleware in priority order
	ListMiddleware() []*Middleware
}
