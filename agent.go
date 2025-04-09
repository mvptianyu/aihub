/*
@Project: aihub
@Module: core
@File : agent.go
*/
package aihub

import (
	"context"
	"errors"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

type agent struct {
	cfg           *AgentConfig
	memory        IMemory
	toolFunctions []ToolFunction

	lock sync.RWMutex
}

func newAgent(cfg *AgentConfig) (IAgent, error) {
	if err := cfg.AutoFix(); err != nil {
		return nil, err
	}

	ag := &agent{
		cfg:    cfg,
		memory: newMemory(&cfg.AgentRuntimeCfg),
	}

	// 初始化系统提示词
	if err := ag.initSystem(); err != nil {
		return nil, err
	}
	return ag, nil
}

func (a *agent) initSystem() error {
	if a.cfg.SystemPrompt == "" {
		return nil
	}

	sysMsg := &Message{
		Role:    MessageRoleSystem,
		Content: a.cfg.SystemPrompt,
	}

	options := a.NewRunOptions()
	sysMsg.Content = options.FixMessageContent(MessageRoleSystem, sysMsg.Content)
	a.memory.SetSystemMsg(sysMsg)
	return nil
}

// 重置对话
func (a *agent) ResetMemory(ctx context.Context, opts ...RunOptionFunc) error {
	options := a.NewRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	a.memory.Clear(options)
	return nil
}

func (a *agent) Run(ctx context.Context, input string, opts ...RunOptionFunc) (*Message, string, error) {
	providerIns := GetProviderHub().GetProvider(a.cfg.Provider)
	if providerIns == nil {
		return nil, "", ErrConfiguration
	}

	options := a.NewRunOptions()
	for _, opt := range opts {
		opt(options)
	}
	if options.SessionID == "" {
		options.SessionID = uuid.NewV4().String()
	}
	options.Question = input

	newCtx, cancel := context.WithTimeout(ctx, time.Duration(options.RuntimeCfg.RunTimeout)*time.Second)
	defer cancel()

	userMsg := &Message{
		Role:    MessageRoleUser,
		Content: input,
	}
	a.memory.Push(options, userMsg)

	var ret *Message
	var content = ""
	var err error
	var doneCh = make(chan bool)

	go func() {
		defer func() {
			doneCh <- true
		}()

		for {
			// 超过最大步数跳出
			if options.CheckStepQuit() {
				err = ErrChatCompletionOverMaxStep
				return
			}

			req := &CreateChatCompletionReq{
				Messages:         a.memory.GetLatest(options),
				Tools:            a.getToolCfg(),
				MaxTokens:        options.RuntimeCfg.MaxTokens,
				FrequencyPenalty: options.RuntimeCfg.FrequencyPenalty,
				PresencePenalty:  options.RuntimeCfg.PresencePenalty,
				Temperature:      options.RuntimeCfg.Temperature,
			}

			// 结束词规则
			if options.RuntimeCfg.StopWords != "" {
				req.Stop = options.RuntimeCfg.StopWords
			}

			// SysPrompt实时替换，例如Context
			sysMsg := req.Messages[0]
			if sysMsg.Role == MessageRoleSystem {
				sysMsg.Content = options.FixMessageContent(MessageRoleSystem, sysMsg.Content)
			}

			rsp, err1 := providerIns.CreateChatCompletion(newCtx, req)
			if err1 != nil {
				err = err1
				return
			}

			if rsp.Error != nil {
				err = errors.New(rsp.Error.Message)
				return
			}

			choice := rsp.Choices[0]
			a.memory.Push(options, choice.Message)

			switch choice.FinishReason {
			case ChatCompletionRspFinishReasonToolCalls:
				// 处理tool调用
				toolMsgs, err1 := a.processToolCalls(newCtx, choice.Message.ToolCalls, options)
				if err1 != nil {
					err = err1
					return
				}

				options.AddStep(choice.Message.ToolCalls, toolMsgs)
				a.memory.Push(options, toolMsgs...)
			default:
				ret = choice.Message
				content = choice.Message.Content
				options.FinalAnswer = choice.Message.Content

				// 修正回复
				content = options.FixMessageContent(MessageRoleAssistant, choice.Message.Content)
				return
			}
		}
	}()

	select {
	case <-newCtx.Done():
		err = ErrAgentRunTimeout
		return ret, content, err
	case <-doneCh:
		return ret, content, err
	}
}

func (a *agent) RunStream(ctx context.Context, input string, opts ...RunOptionFunc) (<-chan Message, <-chan string, <-chan error) {
	// TODO implement me
	panic("implement me")
}

func (m *agent) GetToolFunctions() []ToolFunction {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.toolFunctions != nil {
		return m.toolFunctions
	}

	m.toolFunctions = make([]ToolFunction, 0)
	if m.cfg.Mcps != nil && len(m.cfg.Mcps) > 0 {
		m.toolFunctions = append(m.toolFunctions, GetMCPHub().GetToolFunctions(m.cfg.Mcps...)...)
	}
	if m.cfg.Tools != nil && len(m.cfg.Tools) > 0 {
		m.toolFunctions = append(m.toolFunctions, GetToolHub().GetToolFunctions(m.cfg.Tools...)...)
	}

	return m.toolFunctions
}

func (a *agent) NewRunOptions() *RunOptions {
	options := &RunOptions{
		RuntimeCfg: a.cfg.AgentRuntimeCfg,
		Tools:      a.GetToolFunctions(),
	}
	return options
}

func (m *agent) getToolCfg() []*Tool {
	toolFunctions := m.GetToolFunctions()

	ret := make([]*Tool, 0)
	for _, item := range toolFunctions {
		ret = append(ret, &Tool{
			Type:     ToolTypeFunction,
			Function: item,
		})
	}
	return ret
}

// processToolCalls 处理本步骤toolCalls
func (a *agent) processToolCalls(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) (toolMsgs []*Message, err error) {
	middlewares := make([]IMiddleware, 0)
	if a.cfg.Middlewares != nil {
		middlewares = GetMiddlewareHub().GetMiddleware(a.cfg.Middlewares...)
	}

	// 前处理
	for i := 0; i < len(middlewares); i++ {
		middleware := middlewares[i]
		if err = middleware.BeforeProcessing(ctx, toolCalls, opts); err != nil {
			return nil, err
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(toolCalls))
	toolMsgs = make([]*Message, len(toolCalls))
	for i := 0; i < len(toolCalls); i++ {
		toolCall := toolCalls[i]
		toolMsgs[i] = &Message{
			Role:         MessageRoleTool,
			ToolCallID:   toolCall.Id,
			MultiContent: make([]*MessageContentPart, 0),
		}

		go func(i int, toolCall *MessageToolCall) {
			defer wg.Done()

			a.invokeToolCall(ctx, toolCall, toolMsgs[i])
		}(i, toolCall)
	}
	wg.Wait()

	// 后处理
	for j := len(middlewares) - 1; j >= 0; j-- {
		middleware := middlewares[j]
		if err = middleware.AfterProcessing(ctx, toolCalls, opts); err != nil {
			return nil, err
		}
	}

	return
}

// invokeToolCall 处理本步骤toolCall
func (m *agent) invokeToolCall(ctx context.Context, toolCall *MessageToolCall, output *Message) {
	// 1.MCP调用
	err := GetMCPHub().ProxyCall(ctx, toolCall.Function.Name, toolCall.Function.Arguments, output)
	if errors.Is(err, ErrCallNameNotMatch) {
		// 2.ToolCall本地调用
		err = GetToolHub().ProxyCall(ctx, toolCall.Function.Name, toolCall.Function.Arguments, output)
	}

	if err != nil {
		output.Content = err.Error()
	}
}
