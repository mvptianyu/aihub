package aihub

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/mvptianyu/aihub/ssestream"
	"io"
	"sync"
	"time"
)

type agent struct {
	cfg           *AgentConfig
	memory        IMemory
	toolFunctions []ToolFunction
	lock          sync.RWMutex
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
		Content: opts.UpdateSystemPrompt(a.cfg.SystemPrompt),
	}
}

func (a *agent) GetBriefInfo() BriefInfo {
	return a.cfg.BriefInfo
}

// ResetMemory 重置对话记录
func (a *agent) ResetMemory(ctx context.Context, opts ...RunOptionFunc) error {
	options := a.newRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	a.memory.Clear(options)
	return nil
}

func (a *agent) Run(ctx context.Context, input string, opts ...RunOptionFunc) (ret *Response) {
	ret = &Response{}
	LLMIns := GetLLMHub().GetLLM(a.cfg.LLM)
	if LLMIns == nil {
		ret.Err = ErrConfiguration
		return
	}

	options := a.newRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	ret.Session = options.Session
	ctx = ContextWithSession(ctx, options.Session) // 绑定重设ctx
	newCtx, cancel := context.WithTimeout(ctx, time.Duration(options.RuntimeCfg.RunTimeout)*time.Second)
	defer cancel()

	userMsg := &Message{
		Role:    MessageRoleUser,
		Content: input,
	}
	a.memory.Push(options, userMsg)

	var doneCh = make(chan bool)
	var endStep = &RunStep{
		StepType: StepType_End,
		State:    RunState_Idle,
	}
	options.AddStep(&RunStep{
		Question: input,
		StepType: StepType_Start,
		State:    RunState_Succeed,
	})

	go func() {
		defer func() {
			doneCh <- true
		}()

		for {
			// 超过最大步数跳出
			if options.CheckStepQuit() {
				ret.Err = ErrChatCompletionOverMaxStep
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

			rsp, err1 := LLMIns.CreateChatCompletion(newCtx, req)
			if err1 != nil {
				ret.Err = err1
				return
			}

			if rsp.Error != nil {
				ret.Err = errors.New(rsp.Error.Message)
				return
			}

			choice := rsp.Choices[0]
			a.memory.Push(options, choice.Message)

			switch choice.FinishReason {
			case ChatCompletionRspFinishReasonToolCalls:
				// 处理tool调用
				toolMsgs, err1 := a.processToolCalls(newCtx, choice.Message, options)
				if err1 != nil {
					ret.Err = err1
					return
				}
				a.memory.Push(options, toolMsgs...)
			default:
				ret.Message = choice.Message
				endStep.Result = choice.Message.Content
				return
			}
		}
	}()

	select {
	case <-newCtx.Done():
		ret.Err = ErrAgentRunTimeout
	case <-doneCh:
	}

	if ret.Err != nil {
		endStep.State = RunState_Failed
	}
	options.AddStep(endStep)
	ret.Content = options.RenderFinalAnswer()
	return ret
}

func (a *agent) RunStream(ctx context.Context, input string, opts ...RunOptionFunc) (stream *ssestream.StreamReader[Response]) {
	var err error
	r, w := io.Pipe()
	writer := ssestream.NewStreamWriter[Response](ssestream.NewEncoder(w), ctx)
	stream = ssestream.NewStreamReader[Response](ssestream.NewDecoder(r), err)

	go func() {
		rsp := a.Run(ctx, input, opts...)
		if rsp.Err != nil {
			writer.Append(&Response{
				Err: rsp.Err,
			})
			time.Sleep(30 * time.Millisecond)
			writer.Close()
			return
		}

		// 成功
		runes := []rune(rsp.Content)
		length := len(runes)
		for i := 0; i < length; i += 4 {
			end := i + 4
			if end > length {
				end = length
			}
			block := &Response{}
			block.Content = string(runes[i:end])
			writer.Append(block)
			time.Sleep(25 * time.Millisecond)
		}
		writer.Close()
	}()
	return
}

func (a *agent) GetToolFunctions() []ToolFunction {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.toolFunctions != nil {
		return a.toolFunctions
	}

	a.toolFunctions = make([]ToolFunction, 0)
	// 全局tool白名单
	if a.cfg.Tools == nil || len(a.cfg.Tools) <= 0 {
		return a.toolFunctions
	}

	if a.cfg.Mcps != nil && len(a.cfg.Mcps) > 0 {
		a.toolFunctions = append(a.toolFunctions, GetMCPHub().GetToolFunctions(a.cfg.Mcps, a.cfg.Tools)...)
	}

	if a.cfg.Tools != nil && len(a.cfg.Tools) > 0 {
		a.toolFunctions = append(a.toolFunctions, GetToolHub().GetToolFunctions(a.cfg.Tools...)...)
	}

	return a.toolFunctions
}

func (a *agent) newRunOptions() *RunOptions {
	options := &RunOptions{
		RuntimeCfg: a.cfg.AgentRuntimeCfg,
		Tools:      a.getRelatedToolBriefInfos(),
		Session:    newSession(a.cfg.SessionData),
	}
	return options
}

func (a *agent) getRelatedToolBriefInfos() []BriefInfo {
	toolFunctions := a.GetToolFunctions()

	ret := make([]BriefInfo, 0)
	for _, item := range toolFunctions {
		ret = append(ret, item.BriefInfo)
	}
	return ret
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
func (a *agent) processToolCalls(ctx context.Context, req *Message, opts *RunOptions) (rsp []*Message, err error) {
	middlewares := make([]IMiddleware, 0)
	if a.cfg.Middlewares != nil {
		middlewares = GetMiddlewareHub().GetMiddleware(a.cfg.Middlewares...)
	}

	// 前处理
	for i := 0; i < len(middlewares); i++ {
		middleware := middlewares[i]
		if err = middleware.BeforeProcessing(ctx, req, rsp, opts); err != nil {
			return nil, err
		}
	}

	cnt := len(req.ToolCalls)
	wg := sync.WaitGroup{}
	wg.Add(cnt)
	rsp = make([]*Message, cnt)
	steps := make([]*RunStep, cnt)
	for i := 0; i < cnt; i++ {
		toolCall := req.ToolCalls[i]
		rsp[i] = &Message{
			Role:         MessageRoleTool,
			ToolCallID:   toolCall.Id,
			MultiContent: make([]*MessageContentPart, 0),
		}

		steps[i] = &RunStep{
			Action:   toolCall.Function.Name,
			Question: toolCall.Function.Arguments,
			State:    RunState_Running,
			StepType: StepType_Tool,
		}
		opts.AddStep(steps[i])

		go func(i int, toolCall *MessageToolCall) {
			defer wg.Done()

			err1 := a.InvokeToolCall(ctx, toolCall.Function.Name, toolCall.Function.Arguments, rsp[i])
			steps[i].Result = rsp[i].Content
			steps[i].State = RunState_Failed
			if err1 == nil {
				steps[i].State = RunState_Succeed
			}
		}(i, toolCall)
	}
	wg.Wait()

	// 判断加入SessionKey
	for i := 0; i < cnt; i++ {
		toolCall := req.ToolCalls[i]
		tmpInput := &ToolInputBase{}
		json.Unmarshal([]byte(toolCall.Function.Arguments), tmpInput)
		if tmpInput.GetRawSession() != "" {
			opts.SetSessionData(tmpInput.GetRawSession(), rsp[i].Content)
		}
	}

	// 后处理
	for j := len(middlewares) - 1; j >= 0; j-- {
		middleware := middlewares[j]
		if err = middleware.AfterProcessing(ctx, req, rsp, opts); err != nil {
			return nil, err
		}
	}

	return
}

// InvokeToolCall 处理本步骤toolCall
func (a *agent) InvokeToolCall(ctx context.Context, name string, args string, output *Message) (err error) {
	// 1.MCP调用
	err = GetMCPHub().ProxyCall(ctx, name, args, output)
	if errors.Is(err, ErrCallNameNotMatch) {
		// 2.ToolCall本地调用
		err = GetToolHub().ProxyCall(ctx, name, args, output)
	}

	if err != nil {
		output.Content = err.Error()
		return
	}

	if output.Content == "" {
		err = ErrToolCallResponseEmpty
		output.Content = err.Error()
	}

	return
}
