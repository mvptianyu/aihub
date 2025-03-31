/*
@Project: aihub
@Module: core
@File : agent.go
*/
package core

import (
	"context"
	"errors"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

type Agent struct {
	cfg      *AgentConfig
	provider IProvider
	history  *memory
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
		history:  NewMemory(cfg.MaxStoreHistory, cfg.HistoryTimeout),
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

	options := a.NewRunOptions()
	ctx := context.Background()
	sysMsg.Content = options.FixMessageContent(MessageRoleSystem, sysMsg.Content)
	req := &CreateChatCompletionReq{
		Messages:         []*Message{sysMsg},
		Tools:            a.toolMgr.GetToolCfg(),
		MaxTokens:        a.cfg.MaxTokens,
		FrequencyPenalty: a.cfg.FrequencyPenalty,
		PresencePenalty:  a.cfg.PresencePenalty,
		Temperature:      a.cfg.Temperature,
	}

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
func (a *Agent) ResetHistory(ctx context.Context, opts ...RunOptionFunc) error {
	options := a.NewRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	a.history.Clear(options)
	return nil
}

func (a *Agent) NewRunOptions() *RunOptions {
	options := &RunOptions{
		RuntimeCfg: a.cfg.AgentRuntimeCfg,
		Tools:      a.toolMgr.GetToolDefinition(),
		SessionID:  uuid.NewV4().String(),
		CreateTime: time.Now().Unix(),
	}
	options.RuntimeCfg = a.cfg.AgentRuntimeCfg
	options.Tools = a.toolMgr.GetToolDefinition()
	return options
}

func (a *Agent) Run(ctx context.Context, input string, opts ...RunOptionFunc) (*Message, string, error) {
	options := a.NewRunOptions()
	for _, opt := range opts {
		opt(options)
	}
	if options.SessionID == "" {
		options.SessionID = uuid.NewV4().String()
	}

	userMsg := &Message{
		Role:    MessageRoleUser,
		Content: input,
	}

	a.history.Push(options, userMsg)
	options.SetQuestion(input)

	for {
		// 超过最大步数跳出
		if options.CheckStepQuit() {
			return nil, "", ErrChatCompletionOverMaxStep
		}

		req := &CreateChatCompletionReq{
			Messages:         a.history.GetLatest(options),
			Tools:            a.toolMgr.GetToolCfg(),
			MaxTokens:        options.RuntimeCfg.MaxTokens,
			FrequencyPenalty: options.RuntimeCfg.FrequencyPenalty,
			PresencePenalty:  options.RuntimeCfg.PresencePenalty,
			Temperature:      options.RuntimeCfg.Temperature,
		}

		// 结束词规则
		if options.RuntimeCfg.StopWords != "" {
			req.Stop = options.RuntimeCfg.StopWords
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
		a.history.Push(options, choice.Message)

		switch choice.FinishReason {
		case ChatCompletionRspFinishReasonToolCalls:
			// 处理tool调用
			// toolMsgs, err1 := a.processToolCalls(ctx, choice.Message.ToolCalls, options)
			toolMsgs, err1 := a.toolMgr.ProcessToolCalls(ctx, choice.Message.ToolCalls, options)
			if err1 != nil {
				return nil, "", err1
			}

			options.AddStep(choice.Message.ToolCalls, toolMsgs)
			a.history.Push(options, toolMsgs...)
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
