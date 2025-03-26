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
	tools    map[string]*Tool
	history  *history

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

	go ag.initTools()
	return ag
}

func (a *Agent) initTools() {
	if a.cfg.Tools == nil || len(a.cfg.Tools) == 0 {
		return
	}

	for _, toolFunction := range a.cfg.Tools {
		if toolFunction.Parameters == nil {
			toolFunction.Parameters = &ToolFunctionParameters{
				Type: ToolFunctionParametersTypeText,
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
}

// 重置对话
func (a *Agent) Refresh(ctx context.Context) error {
	a.history.Clear()
	return nil
}

func (a *Agent) Run(ctx context.Context, input string) (*Message, string, error) {
	userMsg := &Message{
		Role:    MessageRoleUser,
		Content: input,
	}

	a.history.Push(userMsg)

	for {
		req := &CreateChatCompletionReq{
			Messages:         a.history.GetAll(*a.cfg.MaxUseHistory, a.cfg.SystemPrompt != ""),
			Tools:            a.ListTool(),
			MaxTokens:        *a.cfg.MaxTokens,
			FrequencyPenalty: *a.cfg.FrequencyPenalty,
			PresencePenalty:  *a.cfg.PresencePenalty,
			Temperature:      *a.cfg.Temperature,
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

		// 处理tool
		if choice.FinishReason == ChatCompletionRspFinishReasonToolCalls {
			toolMsgs, err1 := a.processToolCalls(choice.Message.ToolCalls)
			if err1 != nil {
				return nil, "", err1
			}

			// todo: 再发Chat请求
			a.history.Push(toolMsgs...)
		}
	}

	return choice.Message, choice.Message.Content, err

}

func (a *Agent) RunStream(ctx context.Context, input string) (<-chan Message, <-chan string, <-chan error) {
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

func (a *Agent) processToolCalls(toolCalls []*MessageToolCall) (toolMsgs []*Message, err error) {
	wg := sync.WaitGroup{}
	wg.Add(len(toolCalls))
	toolMsgs = make([]*Message, 0, len(toolCalls))
	for i := 0; i < len(toolCalls); i++ {
		toolCall := toolCalls[i]
		toolMsgs[i] = &Message{
			Role:       MessageRoleTool,
			ToolCallID: toolCall.Id,
		}

		go func(i int, toolCall *MessageToolCall) {
			defer wg.Done()

			output, err1 := a.runToolCall(toolCall)
			if err1 != nil {
				err = err1
				return
			}

			toolMsgs[i].Content = output
		}(i, toolCall)
	}
	wg.Wait()
	return
}

func (a *Agent) runToolCall(tool *MessageToolCall) (output string, err error) {
	output = ""
	return
}
