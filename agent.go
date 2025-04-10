package aihub

import (
	"context"
	"encoding/json"
	"errors"
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
	return ag, nil
}

// getSystemMsg 获取系统消息
func (a *agent) getSystemMsg(opts *RunOptions) *Message {
	if a.cfg.SystemPrompt == "" {
		return nil
	}

	return &Message{
		Role:    MessageRoleSystem,
		Content: opts.FixMessageContent(MessageRoleSystem, a.cfg.SystemPrompt),
	}
}

// ResetMemory 重置对话记录
func (a *agent) ResetMemory(ctx context.Context, opts ...RunOptionFunc) error {
	options := a.NewRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	a.memory.Clear(options)
	return nil
}

func (a *agent) Run(ctx context.Context, input string, opts ...RunOptionFunc) (*Message, string, ISession, error) {
	providerIns := GetProviderHub().GetProvider(a.cfg.Provider)
	if providerIns == nil {
		return nil, "", nil, ErrConfiguration
	}

	options := a.NewRunOptions()
	for _, opt := range opts {
		opt(options)
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

			messages := make([]*Message, 0)
			messages = append(messages, a.getSystemMsg(options))        // system
			messages = append(messages, a.memory.GetLatest(options)...) // latest N

			req := &CreateChatCompletionReq{
				Messages:         messages,
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
	case <-doneCh:
	}

	return ret, content, options.Session, err
}

func (a *agent) RunStream(ctx context.Context, input string, opts ...RunOptionFunc) (<-chan Message, <-chan string, <-chan error) {
	// TODO implement me
	panic("implement me")
}

func (a *agent) GetToolFunctions() []ToolFunction {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.toolFunctions != nil {
		return a.toolFunctions
	}

	a.toolFunctions = make([]ToolFunction, 0)
	if a.cfg.Mcps != nil && len(a.cfg.Mcps) > 0 {
		a.toolFunctions = append(a.toolFunctions, GetMCPHub().GetToolFunctions(a.cfg.Mcps...)...)
	}
	if a.cfg.Tools != nil && len(a.cfg.Tools) > 0 {
		a.toolFunctions = append(a.toolFunctions, GetToolHub().GetToolFunctions(a.cfg.Tools...)...)
	}

	return a.toolFunctions
}

func (a *agent) NewRunOptions() *RunOptions {
	options := &RunOptions{
		RuntimeCfg: a.cfg.AgentRuntimeCfg,
		Tools:      a.GetToolFunctions(),
		Session:    newSession(a.cfg.SessionData),
	}
	return options
}

func (a *agent) getToolCfg() []*Tool {
	toolFunctions := a.GetToolFunctions()

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

	cnt := len(toolCalls)
	wg := sync.WaitGroup{}
	wg.Add(cnt)
	toolMsgs = make([]*Message, len(toolCalls))
	for i := 0; i < cnt; i++ {
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

	// 判断加入SessionKey
	for i := 0; i < cnt; i++ {
		toolCall := toolCalls[i]
		tmpInput := &ToolInputBase{}
		json.Unmarshal([]byte(toolCall.Function.Arguments), tmpInput)
		if tmpInput.GetRawSession() != "" {
			opts.SetSessionData(tmpInput.GetRawSession(), toolMsgs[i].Content)
		}
	}

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
func (a *agent) invokeToolCall(ctx context.Context, toolCall *MessageToolCall, output *Message) {
	// 1.MCP调用
	err := GetMCPHub().ProxyCall(ctx, toolCall.Function.Name, toolCall.Function.Arguments, output)
	if errors.Is(err, ErrCallNameNotMatch) {
		// 2.ToolCall本地调用
		err = GetToolHub().ProxyCall(ctx, toolCall.Function.Name, toolCall.Function.Arguments, output)
	}

	if err != nil {
		output.Content = err.Error()
	}

	if output.Content == "" {
		output.Content = ErrToolCallResponseEmpty.Error()
	}
}
