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
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Agent struct {
	cfg         *AgentConfig
	provider    IProvider
	memory      IMemory
	toolMgr     IToolManager
	middlewares []IMiddleware

	lock sync.RWMutex
}

func NewAgent(cfg *AgentConfig, toolDelegate interface{}) IAgent {
	if err := cfg.AutoFix(); err != nil {
		panic(err)
	}

	ag := &Agent{
		cfg:      cfg,
		provider: NewProvider(&cfg.Provider),
		memory:   NewMemory(&cfg.AgentRuntimeCfg),
		toolMgr:  NewToolManager(&cfg.AgentRuntimeCfg),
	}

	// 先初始化MCP、本地工具
	if err := ag.toolMgr.RegisterMCPFunc(); err != nil {
		panic(err)
	}
	if err := ag.toolMgr.RegisterToolFunc(toolDelegate); err != nil {
		panic(err)
	}

	// 再初始化系统提示词
	if err := ag.initSystem(); err != nil {
		panic(err)
	}
	return ag
}

// NewAgentWithYaml 从配置读取
func NewAgentWithYamlData(yamlData []byte, toolDelegate interface{}) IAgent {
	cfg := &AgentConfig{}
	cfg.Mcps = make([]string, 0)
	if err := yaml.Unmarshal(yamlData, cfg); err != nil {
		log.Fatalf("Error Unmarshal YAML data: %s => %v\n", string(yamlData), err)
		return nil
	}

	return NewAgent(cfg, toolDelegate)
}

// NewAgentWithYamlFile 从配置文件读取
func NewAgentWithYamlFile(yamlFile string, toolDelegate interface{}) IAgent {
	// 读取 YAML 文件内容
	yamlData, err := os.ReadFile(filepath.Clean(yamlFile))
	if err != nil {
		log.Fatalf("Error reading YAML file: %s => %v\n", yamlFile, err)
		return nil
	}

	return NewAgentWithYamlData(yamlData, toolDelegate)
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
	sysMsg.Content = options.FixMessageContent(MessageRoleSystem, sysMsg.Content)
	a.memory.SetSystemMsg(sysMsg)
	return nil
}

// 重置对话
func (a *Agent) ResetMemory(ctx context.Context, opts ...RunOptionFunc) error {
	options := a.NewRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	a.memory.Clear(options)
	return nil
}

func (a *Agent) NewRunOptions() *RunOptions {
	options := &RunOptions{
		RuntimeCfg: a.cfg.AgentRuntimeCfg,
		Tools:      a.toolMgr.GetToolDefinition(),
		DoneCh:     make(chan bool),
	}
	return options
}

func (a *Agent) RegisterMiddleware(middleware ...IMiddleware) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.middlewares = append(a.middlewares, middleware...)
}

func (a *Agent) Run(ctx context.Context, input string, opts ...RunOptionFunc) (*Message, string, error) {
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

	go func() {
		defer func() {
			options.DoneCh <- true
		}()

		for {
			// 超过最大步数跳出
			if options.CheckStepQuit() {
				err = ErrChatCompletionOverMaxStep
				return
			}

			req := &CreateChatCompletionReq{
				Messages:         a.memory.GetLatest(options),
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

			rsp, err1 := a.provider.CreateChatCompletion(newCtx, req)
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
	case <-options.DoneCh:
		return ret, content, err
	}
}

func (a *Agent) RunStream(ctx context.Context, input string, opts ...RunOptionFunc) (<-chan Message, <-chan string, <-chan error) {
	// TODO implement me
	panic("implement me")
}

// processToolCalls 处理本步骤tookcalls
func (a *Agent) processToolCalls(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) (toolMsgs []*Message, err error) {
	// 前处理
	for i := 0; i < len(a.middlewares); i++ {
		middleware := a.middlewares[i]
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

			err1 := a.toolMgr.InvokeToolFunc(ctx, toolCall, toolMsgs[i])
			if err1 != nil {
				err = err1
				return
			}
		}(i, toolCall)
	}
	wg.Wait()

	// 后处理
	for j := len(a.middlewares) - 1; j >= 0; j-- {
		middleware := a.middlewares[j]
		if err = middleware.AfterProcessing(ctx, toolCalls, opts); err != nil {
			return nil, err
		}
	}

	return
}
