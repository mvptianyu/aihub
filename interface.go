/*
@Project: aihub
@Module: core
@File : interface.go
*/
package aihub

import (
	"context"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// IProvider 模型提供商相关能力
type IProvider interface {
	// CreateChatCompletion 创建Chat
	CreateChatCompletion(ctx context.Context, request *CreateChatCompletionReq) (response *CreateChatCompletionRsp, err error)
	// CreateChatCompletionStream 创建Chat以及stream返回
	CreateChatCompletionStream(ctx context.Context, request *CreateChatCompletionReq) (stream *Stream[CreateChatCompletionRsp])
}

// IAgent interface defines the core capabilities required for an agent
type IAgent interface {
	// Run executes the agent's main loop with the given input until a stop condition is met
	Run(ctx context.Context, input string, opts ...RunOptionFunc) (*Message, string, error)
	// RunStream supports a streaming channel from a provider
	RunStream(ctx context.Context, input string, opts ...RunOptionFunc) (<-chan Message, <-chan string, <-chan error)
	// ResetMemory 重置会话记忆
	ResetMemory(ctx context.Context, opts ...RunOptionFunc) error
	// GetToolFunctions 获取工具配置
	GetToolFunctions() []ToolFunction
}

// 会话记录
type IMemory interface {
	// GetSystemMsg 获取会话系统消息
	GetSystemMsg() *Message
	// SetSystemMsg 设置会话系统消息
	SetSystemMsg(msg *Message)
	// Push 塞入会话消息记录
	Push(opts *RunOptions, msg ...*Message)
	// GetLatest 获取最近会话消息记录
	GetLatest(opts *RunOptions) []*Message
	// Clear 清理指定消息记录
	Clear(opts *RunOptions)
}

// 调用拦截器
type IMiddleware interface {
	// BeforeProcessing 前处理
	BeforeProcessing(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) error
	// AfterProcessing 后处理
	AfterProcessing(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) error
}

// --------------
type IMiddlewareHub interface {
	GetMiddleware(names ...string) []IMiddleware
	DelMiddleware(names ...string) error
	SetMiddleware(middlewares ...IMiddleware) error
}

type IToolHub interface {
	GetToolFunctions(names ...string) []ToolFunction
	GetTool(names ...string) []ToolEntry
	DelTool(names ...string) error
	SetTool(objs ...ToolEntry) error
	ProxyCall(ctx context.Context, name string, input string, output *Message) (err error)
}

type IMCPHub interface {
	GetClient(addrs ...string) []*client.SSEMCPClient
	DelClient(addrs ...string) error
	SetClient(addrs ...string) error
	ProxyCall(ctx context.Context, name string, input string, output *Message) (rsp *mcp.CallToolResult, err error)
	GetToolFunctions(addrs ...string) []ToolFunction
}

type IProviderHub interface {
	GetProviderList(names ...string) []IProvider
	GetProvider(name string) IProvider
	DelProvider(name string) error
	SetProvider(cfg *ProviderConfig) (IProvider, error)
	SetProviderByYamlData(yamlData []byte) (IProvider, error)
	SetProviderByYamlFile(yamlFile string) (IProvider, error)
}

type IAgentHub interface {
	GetAgentList(names ...string) []IAgent
	GetAgent(name string) IAgent
	DelAgent(name string) error
	SetAgent(cfg *AgentConfig) (IAgent, error)
	SetAgentByYamlData(yamlData []byte) (IAgent, error)
	SetAgentByYamlFile(yamlFile string) (IAgent, error)
}
