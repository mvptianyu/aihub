/*
@Project: aihub
@Module: core
@File : interface.go
*/
package core

import (
	"context"
)

// IProvider 模型提供商相关能力
type IProvider interface {
	CreateChatCompletion(ctx context.Context, request *CreateChatCompletionReq) (response *CreateChatCompletionRsp, err error)

	CreateChatCompletionStream(ctx context.Context, request *CreateChatCompletionReq) (stream *Stream[CreateChatCompletionRsp])
}

// IAgent interface defines the core capabilities required for an agent
type IAgent interface {
	// Run executes the agent's main loop with the given input until a stop condition is met
	Init(router ToolFuncRouter)

	// Run executes the agent's main loop with the given input until a stop condition is met
	Run(ctx context.Context, input string) (*Message, string, error)

	// RunStream supports a streaming channel from a provider
	RunStream(ctx context.Context, input string) (<-chan Message, <-chan string, <-chan error)

	// Run executes the agent's main loop with the given input until a stop condition is met
	ResetHistory() error

	// GetTool returns the tool
	GetTool(name string) (*Tool, bool)

	// RegisterTool adds a new tool to the agent's capabilities
	RegisterTool(tool *Tool) error

	// RemoveTool removes a tool by name
	RemoveTool(name string) error

	// ListTool returns the current set of available tools
	ListTool() []*Tool
}
