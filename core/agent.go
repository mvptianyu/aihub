/*
@Project: aihub
@Module: core
@File : agent.go
*/
package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mvptianyu/aihub/jsonschema"
	"github.com/tidwall/gjson"
	"strings"
	"sync"
)

const (
	defaultPromptReplaceContext = "{{context}}"
	defaultPromptReplaceTools   = "{{tools}}"
	defaultPromptReplaceMCP     = "{{mcp}}"
)

type Agent struct {
	cfg        *AgentConfig
	provider   IProvider
	tools      map[string]*Tool
	history    *history
	toolRouter ToolFuncRouter
	inited     bool

	lock sync.RWMutex
}

func NewAgent(cfg *AgentConfig) IAgent {
	if err := cfg.AutoFix(); err != nil {
		panic(err)
	}

	ag := &Agent{
		cfg:      cfg,
		provider: NewProvider(&cfg.Provider),
		tools:    make(map[string]*Tool),
		history:  NewHistory(*cfg.MaxStoreHistory),
	}

	return ag
}

func (a *Agent) Init(router ToolFuncRouter) IAgent {
	a.toolRouter = router

	if a.inited {
		return a
	}

	// 先初始化工具
	if err := a.initTools(); err != nil {
		panic(err)
	}

	// 再初始化系统提示词
	if err := a.initSystem(); err != nil {
		panic(err)
	}

	a.inited = true
	return a
}

func (a *Agent) initSystem() error {
	if a.cfg.SystemPrompt == "" {
		return nil
	}

	sysMsg := &Message{
		Role:    MessageRoleSystem,
		Content: a.cfg.SystemPrompt,
	}

	ctx := context.Background()
	req := &CreateChatCompletionReq{
		Messages:         []*Message{sysMsg},
		Tools:            a.ListTool(),
		MaxTokens:        *a.cfg.MaxTokens,
		FrequencyPenalty: *a.cfg.FrequencyPenalty,
		PresencePenalty:  *a.cfg.PresencePenalty,
		Temperature:      *a.cfg.Temperature,
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

func (a *Agent) initTools() error {
	if a.cfg.Tools == nil || len(a.cfg.Tools) == 0 {
		return nil
	}

	for _, toolFunction := range a.cfg.Tools {
		if toolFunction.Name == "" {
			continue
		}

		if toolFunction.Parameters == nil {
			toolFunction.Parameters = &jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					ToolFunctionDefaultParam: {
						Type:        jsonschema.String,
						Description: "tools's input parameter",
					},
				},
				Required: []string{ToolFunctionDefaultParam},
			}
		}
		if toolFunction.Description == "" {
			toolFunction.Description = toolFunction.Name
		}

		a.RegisterTool(&Tool{
			Type:     ToolTypeFunction,
			Function: toolFunction,
		})
	}

	if len(a.ListTool()) > 0 && a.toolRouter == nil {
		return ErrToolRouterEmpty
	}
	return nil
}

// 重置对话
func (a *Agent) ResetHistory() error {
	a.history.Clear()
	return nil
}

func (a *Agent) Run(ctx context.Context, input string, opts ...RunOptionFunc) (*Message, string, error) {
	if !a.inited {
		return nil, "", ErrAgentNotInit
	}

	options := &RunOptions{}
	for _, opt := range opts {
		opt(options)
	}

	userMsg := &Message{
		Role:    MessageRoleUser,
		Content: input,
	}

	a.history.Push(userMsg)
	step := 0
	recorder := NewRecorder()
	recorder.SetQuestion(input)

	for {
		// 超过最大步数跳出
		if step > *a.cfg.MaxStepQuit {
			return nil, "", ErrChatCompletionOverMaxStep
		}

		req := &CreateChatCompletionReq{
			Messages:         a.history.GetAll(*a.cfg.MaxUseHistory, a.cfg.SystemPrompt != ""),
			Tools:            a.ListTool(),
			MaxTokens:        *a.cfg.MaxTokens,
			FrequencyPenalty: *a.cfg.FrequencyPenalty,
			PresencePenalty:  *a.cfg.PresencePenalty,
			Temperature:      *a.cfg.Temperature,
		}

		// 上下文实时替换
		if options.Context != "" && req.Messages[0].Role == MessageRoleSystem {
			contextBS, _ := json.Marshal(options.Context)
			req.Messages[0].Content = strings.Replace(req.Messages[0].Content, defaultPromptReplaceContext, string(contextBS), -1)
		}

		// 结束词规则
		if options.StopWords != "" {
			req.Stop = options.StopWords
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

			recorder.AddStep(choice.Message.ToolCalls, toolMsgs)
			a.history.Push(toolMsgs...)
			step++ // 再次请求
		default:
			content := choice.Message.Content
			recorder.SetFinal(choice.Message.Content)

			if options.Debug {
				content = recorder.PrettyPrint()
			}

			if options.Claim != "" {
				content += fmt.Sprintf("\n```ℹ️ %s```", options.Claim)
			}

			return choice.Message, content, nil

		}
	}
}

func (a *Agent) RunStream(ctx context.Context, input string) (<-chan Message, <-chan string, <-chan error) {
	if !a.inited {
		return nil, nil, nil
	}

	// TODO implement me
	panic("implement me")
}

func (a *Agent) GetTool(name string) (*Tool, bool) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	if tool, ok := a.tools[name]; ok {
		return tool, true
	}
	return nil, false
}

func (a *Agent) RegisterTool(tool *Tool) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.tools[tool.Function.Name] = tool
	return nil
}

func (a *Agent) RemoveTool(name string) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if _, ok := a.tools[name]; ok {
		delete(a.tools, name)
	}

	return nil
}

func (a *Agent) ListTool() []*Tool {
	a.lock.RLock()
	defer a.lock.RUnlock()

	ret := make([]*Tool, 0, len(a.tools))
	for _, tool := range a.tools {
		ret = append(ret, tool)
	}
	return ret
}

func (a *Agent) processToolCalls(ctx context.Context, toolCalls []*MessageToolCall) (toolMsgs []*Message, err error) {
	wg := sync.WaitGroup{}
	wg.Add(len(toolCalls))
	toolMsgs = make([]*Message, len(toolCalls))
	for i := 0; i < len(toolCalls); i++ {
		toolCall := toolCalls[i]
		toolMsgs[i] = &Message{
			Role:       MessageRoleTool,
			ToolCallID: toolCall.Id,
		}

		go func(i int, toolCall *MessageToolCall) {
			defer wg.Done()

			args := toolCall.Function.Arguments
			// 如果是未定义schema，用默认ToolFunctionDefaultParam,则拾取拆解作为参数
			if rawArgs := gjson.Get(args, ToolFunctionDefaultParam).String(); rawArgs != "" {
				args = rawArgs
			}

			output, err1 := a.toolRouter(ctx, toolCall.Function.Name, args)
			if err1 != nil {
				err = err1
				return
			}
			bs, _ := json.Marshal(output)
			toolMsgs[i].Content = string(bs)
		}(i, toolCall)
	}
	wg.Wait()
	return
}
