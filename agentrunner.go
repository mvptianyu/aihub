package core

import (
	"context"

	"github.com/mvptianyu/aihub/types"
)

// AgentRunner interface defines the core capabilities required for an agent
type AgentRunner interface {
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
