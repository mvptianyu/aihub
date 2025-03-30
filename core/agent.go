/*
@Project: aihub
@Module: core
@File : agent.go
*/
package core

import (
	"context"
	"errors"
	"sync"
)

type Agent struct {
	cfg      *AgentConfig
	provider IProvider
	history  *history
	toolMgr  *ToolManager

	lock sync.RWMutex
}

func NewAgent(cfg *AgentConfig, toolDelegate interface{}) IAgent {
	if err := cfg.AutoFix(); err != nil {
		panic(err)
	}

	ag := &Agent{
		cfg:      cfg,
		provider: NewProvider(&cfg.Provider),
		history:  NewHistory(*cfg.MaxStoreHistory),
		toolMgr:  NewToolManager(),
	}

	// 先初始化工具
	if err := ag.toolMgr.RegisterToolFunc(toolDelegate, cfg.Tools); err != nil {
		panic(err)
	}

	// 再初始化系统提示词
	if err := ag.initSystem(); err != nil {
		panic(err)
	}
	return ag
}

func (a *Agent) initSystem() error {
	if a.cfg.SystemPrompt == "" {
		return nil
	}

	sysMsg := &Message{
		Role:    MessageRoleSystem,
		Content: a.cfg.SystemPrompt,
	}

	options := NewRunOptions(a)
	ctx := context.Background()
	req := &CreateChatCompletionReq{
		Messages:         []*Message{sysMsg},
		Tools:            a.toolMgr.GetToolCfg(),
		MaxTokens:        *a.cfg.MaxTokens,
		FrequencyPenalty: *a.cfg.FrequencyPenalty,
		PresencePenalty:  *a.cfg.PresencePenalty,
		Temperature:      *a.cfg.Temperature,
	}
	sysMsg.Content = options.FixMessageContent(MessageRoleSystem, sysMsg.Content)

	rsp, err := a.provider.CreateChatCompletion(ctx, req)
	if err != nil {
		return err
	}

	if rsp.Error != nil {
		return errors.New(rsp.Error.Message)
	}

	a.history.SetSystemMsg(sysMsg)
	return nil
}

// 重置对话
func (a *Agent) ResetHistory() error {
	a.history.Clear()
	return nil
}

func (a *Agent) Run(ctx context.Context, input string, opts ...RunOptionFunc) (*Message, string, error) {
	options := NewRunOptions(a)
	for _, opt := range opts {
		opt(options)
	}

	userMsg := &Message{
		Role:    MessageRoleUser,
		Content: input,
	}

	a.history.Push(userMsg)
	options.SetQuestion(input)

	for {
		// 超过最大步数跳出
		if options.CurStep > *a.cfg.MaxStepQuit {
			return nil, "", ErrChatCompletionOverMaxStep
		}

		req := &CreateChatCompletionReq{
			Messages:         a.history.GetAll(*a.cfg.MaxUseHistory, a.cfg.SystemPrompt != ""),
			Tools:            a.toolMgr.GetToolCfg(),
			MaxTokens:        *a.cfg.MaxTokens,
			FrequencyPenalty: *a.cfg.FrequencyPenalty,
			PresencePenalty:  *a.cfg.PresencePenalty,
			Temperature:      *a.cfg.Temperature,
		}

		// 结束词规则
		if options.StopWords != "" {
			req.Stop = options.StopWords
		}

		// SysPrompt实时替换
		sysMsg := req.Messages[0]
		if sysMsg.Role == MessageRoleSystem {
			sysMsg.Content = options.FixMessageContent(MessageRoleSystem, sysMsg.Content)
		}

		rsp, err := a.provider.CreateChatCompletion(ctx, req)
		if err != nil {
			return nil, "", err
		}

		if rsp.Error != nil {
			return nil, "", errors.New(rsp.Error.Message)
		}

		choice := rsp.Choices[0]
		a.history.Push(choice.Message)

		switch choice.FinishReason {
		case ChatCompletionRspFinishReasonToolCalls:
			// 处理tool调用
			toolMsgs, err1 := a.processToolCalls(ctx, choice.Message.ToolCalls)
			if err1 != nil {
				return nil, "", err1
			}

			options.AddStep(choice.Message.ToolCalls, toolMsgs)
			a.history.Push(toolMsgs...)
			options.CurStep++ // 再次请求
		default:
			content := choice.Message.Content
			options.SetFinal(choice.Message.Content)

			// 修正回复
			content = options.FixMessageContent(MessageRoleAssistant, choice.Message.Content)
			return choice.Message, content, nil

		}
	}
}

func (a *Agent) RunStream(ctx context.Context, input string) (<-chan Message, <-chan string, <-chan error) {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) processToolCalls(ctx context.Context, toolCalls []*MessageToolCall) (toolMsgs []*Message, err error) {
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

			err1 := a.toolMgr.InvokeToolFunc(ctx, toolCall, toolMsgs[i])
			if err1 != nil {
				err = err1
				return
			}
		}(i, toolCall)
	}
	wg.Wait()
	return
}
