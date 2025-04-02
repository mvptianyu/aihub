/*
@Project: aihub
@Module: core
@File : interface.go
*/
package aihub

import (
	"context"
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
	// RegisterMiddleware 注册中间件
	RegisterMiddleware(middleware ...IMiddleware)
	// ResetMemory 重置会话记忆
	ResetMemory(ctx context.Context, opts ...RunOptionFunc) error
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

// 会话记录
type IToolManager interface {
	// RegisterMCPFunc 注册MCP服务方法
	RegisterMCPFunc() error
	// RegisterToolFunc 注册工具方法
	RegisterToolFunc(delegate interface{}) error
	// GetToolDefinition 获取工具方法定义
	GetToolDefinition() []ToolFunction
	// GetToolCfg 获取工具方法Prompt信息
	GetToolCfg() []*Tool
	// InvokeToolFunc 调用指定方法
	InvokeToolFunc(ctx context.Context, toolCall *MessageToolCall, output *Message) error
}

// 调用拦截器
type IMiddleware interface {
	// BeforeProcessing 前处理
	BeforeProcessing(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) error
	// OnProcessing 处理中
	OnProcessing(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) error
	// AfterProcessing 后处理
	AfterProcessing(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) error
}
