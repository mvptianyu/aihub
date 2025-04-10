package aihub

import (
	"context"
	"github.com/mark3labs/mcp-go/client"
)

// IProvider 模型提供商相关能力
type IProvider interface {
	// CreateChatCompletion 创建Chat
	CreateChatCompletion(ctx context.Context, request *CreateChatCompletionReq) (response *CreateChatCompletionRsp, err error)
	// CreateChatCompletionStream 创建Chat以及stream返回
	CreateChatCompletionStream(ctx context.Context, request *CreateChatCompletionReq) (stream *Stream[CreateChatCompletionRsp])
}

// IAgent 智能体
type IAgent interface {
	// Run 执行Agent请求
	Run(ctx context.Context, input string, opts ...RunOptionFunc) (*Message, string, ISession, error)
	// RunStream 执行Agent请求，支持流式返回（Todo）
	RunStream(ctx context.Context, input string, opts ...RunOptionFunc) (<-chan Message, <-chan string, <-chan error)
	// ResetMemory 重置会话记忆
	ResetMemory(ctx context.Context, opts ...RunOptionFunc) error
	// GetToolFunctions 获取工具配置
	GetToolFunctions() []ToolFunction
}

// IMemory 会话记录
type IMemory interface {
	// Push 塞入会话消息记录
	Push(opts *RunOptions, msg ...*Message)
	// GetLatest 获取最近会话消息记录
	GetLatest(opts *RunOptions) []*Message
	// Clear 清理指定消息记录
	Clear(opts *RunOptions)
}

// ISession 会话session数据
type ISession interface {
	// SetSessionData 设置数据KV
	SetSessionData(key string, value interface{})
	// GetSessionData 获取数据KV
	GetSessionData(key string) interface{}
	// GetAllSessionData 获取所有数据KV
	GetAllSessionData() map[string]interface{}
	// GetSessionID 获取sessionid
	GetSessionID() string
}

// IMiddleware 调用拦截器
type IMiddleware interface {
	// BeforeProcessing 前处理
	BeforeProcessing(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) error
	// AfterProcessing 后处理
	AfterProcessing(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) error
}

// ========HUB定义============

type IMiddlewareHub interface {
	GetAllNameList() []string
	GetMiddleware(names ...string) []IMiddleware
	DelMiddleware(names ...string) error
	SetMiddleware(middlewares ...IMiddleware) error
}

type IToolHub interface {
	GetAllNameList() []string
	GetToolFunctions(names ...string) []ToolFunction
	GetTool(names ...string) []ToolEntry
	DelTool(names ...string) error
	SetTool(objs ...ToolEntry) error
	ProxyCall(ctx context.Context, name string, input string, output *Message) (err error)
	ConvertToOPENAPIConfig() string
}

type IMCPHub interface {
	GetAllNameList() []string
	GetClient(addrs ...string) []*client.SSEMCPClient
	DelClient(addrs ...string) error
	SetClient(addrs ...string) error
	ProxyCall(ctx context.Context, name string, input string, output *Message) (err error)
	GetToolFunctions(addrs []string, names []string) []ToolFunction
	ConvertToOPENAPIConfig() string
}

type IProviderHub interface {
	GetAllNameList() []string
	GetProviderList(names ...string) []IProvider
	GetProvider(name string) IProvider
	DelProvider(name string) error
	SetProvider(cfg *ProviderConfig) (IProvider, error)
	SetProviderByYamlData(yamlData []byte) (IProvider, error)
	SetProviderByYamlFile(yamlFile string) (IProvider, error)
}

type IAgentHub interface {
	GetAllNameList() []string
	GetAgentList(names ...string) []IAgent
	GetAgent(name string) IAgent
	DelAgent(name string) error
	SetAgent(cfg *AgentConfig) (IAgent, error)
	SetAgentByYamlData(yamlData []byte) (IAgent, error)
	SetAgentByYamlFile(yamlFile string) (IAgent, error)
}
