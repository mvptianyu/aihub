/*
@Project: aihub
@Module: core
@File : option.go
*/
package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type RunOptions struct {
	RuntimeCfg  AgentRuntimeCfg // 运行时配置
	Tools       []ToolFunction  // 用到的工具定义
	SessionID   string
	Question    string
	FinalAnswer string
	DoneCh      chan bool

	context interface{} // 可选，上下文信息，例如知识库等
	steps   []*runOptionsStep
	lock    sync.RWMutex
}

type runOptionsStep struct {
	Action      string
	Observation string
}

const (
	defaultPromptReplaceContext = "{{context}}"
	defaultPromptReplaceTools   = "{{tools}}"
)

const prettyCommonTpl = `
**%s：**
'''
%s
'''
`

const prettyStepTpl = `
**第%d步➡️：**
- **执行🏃‍：** 
'''
%s
'''
- **结果✅：** 
'''
%s
'''
`

func (opts *RunOptions) FixMessageContent(role MessageRoleType, content string) string {
	switch role {
	case MessageRoleSystem:
		if opts != nil && opts.context != nil {
			contentBS, _ := json.Marshal(opts.context)
			content = strings.Replace(content, defaultPromptReplaceContext, string(contentBS), -1)
		}
		if opts.Tools != nil && len(opts.Tools) > 0 {
			toolsBS, _ := json.Marshal(opts.Tools)
			content = strings.Replace(content, defaultPromptReplaceTools, string(toolsBS), -1)
		}
	case MessageRoleAssistant:
		if opts.RuntimeCfg.Debug && content != "" {
			content = opts.PrettyPrint()
		}

		if opts.RuntimeCfg.Claim != "" && content != "" {
			content += fmt.Sprintf("\n```ℹ️ %s```", opts.RuntimeCfg.Claim)
		}
	default:
	}
	return content
}

func (opts *RunOptions) CheckStepQuit() bool {
	opts.lock.RLock()
	defer opts.lock.RUnlock()

	if opts.steps == nil {
		opts.steps = make([]*runOptionsStep, 0)
	}

	// 超过最大步数跳出
	return len(opts.steps) > opts.RuntimeCfg.MaxStepQuit
}

func (opts *RunOptions) AddStep(toolCalls []*MessageToolCall, toolMsgs []*Message) {
	opts.lock.Lock()
	defer opts.lock.Unlock()

	if opts.steps == nil {
		opts.steps = make([]*runOptionsStep, 0)
	}

	action := ""
	observation := ""

	for _, toolCall := range toolCalls {
		action += fmt.Sprintf("%s => %s( %s )\n", toolCall.Id, toolCall.Function.Name, toolCall.Function.Arguments)
	}

	for _, toolMsg := range toolMsgs {
		observation += fmt.Sprintf("%s => %s\n", toolMsg.ToolCallID, toolMsg.Content)
	}

	opts.steps = append(opts.steps, &runOptionsStep{
		Action:      strings.TrimRight(action, "\n"),
		Observation: strings.TrimRight(observation, "\n"),
	})
}

func (opts *RunOptions) PrettyPrint() string {
	opts.lock.RLock()
	defer opts.lock.RUnlock()

	output := fmt.Sprintf(prettyCommonTpl, "用户问题🤔", opts.Question)
	if opts.steps != nil {
		for idx, step := range opts.steps {
			output += fmt.Sprintf(prettyStepTpl, idx+1, step.Action, step.Observation)
		}
	}

	if HasMarkdownSyntax(opts.FinalAnswer) {
		output += opts.FinalAnswer
	} else {
		// 最终结果无格式输出才替换
		output += fmt.Sprintf(prettyCommonTpl, "最终结果📤", opts.FinalAnswer)
	}
	return strings.TrimLeft(strings.Replace(output, "'''", "```", -1), "\n")
}

// ------------
type RunOptionFunc func(*RunOptions)

func WithToolFunctions(tools []ToolFunction) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Tools = tools
	}
}

func WithRuntimeCfg(runtimeCfg AgentRuntimeCfg) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.RuntimeCfg = runtimeCfg
	}
}

func WithDebug(debug bool) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.RuntimeCfg.Debug = debug
	}
}

func WithContext(context interface{}) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.context = context
	}
}

func WithSessionID(sessionID string) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.SessionID = sessionID
	}
}

func WithSessionData(sessionData map[string]interface{}) RunOptionFunc {
	return func(opts *RunOptions) {
		if opts.RuntimeCfg.SessionData == nil {
			opts.RuntimeCfg.SessionData = make(map[string]interface{})
		}
		for k, v := range sessionData {
			opts.RuntimeCfg.SessionData[k] = v
		}
	}
}
