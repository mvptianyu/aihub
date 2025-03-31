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
	Run(ctx context.Context, input string, opts ...RunOptionFunc) (*Message, string, error)

	// RunStream supports a streaming channel from a provider
	RunStream(ctx context.Context, input string) (<-chan Message, <-chan string, <-chan error)

	ResetHistory(ctx context.Context, opts ...RunOptionFunc) error
}

// IApprove 授权管理器
type IMiddleware interface {
	// SubmitApplication 提交授权申请
	BeforeProcessing(ctx context.Context, question string, timeout int64, session map[string]interface{}) error

	// SubmitApplication 提交授权申请
	OnProcessing(ctx context.Context, question string, timeout int64, session map[string]interface{}) error

	// SubmitApplication 提交授权申请
	AfterProcessing(ctx context.Context, question string, timeout int64, session map[string]interface{}) error
}
