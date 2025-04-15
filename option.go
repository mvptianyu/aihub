package aihub

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

type RunOptions struct {
	*Session
	RuntimeCfg AgentRuntimeCfg // 运行时配置
	Tools      []BriefInfo     // 用到的关联tool定义
	Agents     []BriefInfo     // 用到的关联Agent定义
	Context    interface{}     // 可选，上下文信息，例如知识库等

	steps []*RunStep
	lock  sync.RWMutex
}

type RunStep struct {
	Action   string    `json:"_action_" yaml:"_action_" description:"该步骤名称" required:"true"`                                     // 该步骤名称
	State    RunState  `json:"_state_" yaml:"_state_" description:"该步骤运作状态(0-默认初始化，1-执行中，2-成功退出，3-失败退出，4-异常终止)" required:"true"` // 该步骤状态：0-初始化（默认），1-执行中，2-成功退出，3-失败退出，4-异常终止
	Question string    `json:"_question_" yaml:"_question_" description:"该步骤结合用户请求和上下文的的请求输入提示词" required:"true"`                // 该步骤需要解决的问题
	Think    string    `json:"_think_" yaml:"_think_" description:"该步骤结合用户请求和上下文的思考概述" required:"true"`                          // 该步骤的推理思考概要
	Result   string    `json:"_result_,omitempty" yaml:"_result_,omitempty" description:"该步骤运行结果文字内容"`                           // 该步骤完成的输出结果
	EndTime  time.Time `json:",omitempty" yaml:",omitempty"`                                                                     // 该步骤完成时间
	StepType StepType  `json:",omitempty" yaml:",omitempty"`                                                                     // 该步骤类别
}

func (r *RunStep) IsEmpty() bool {
	if r == nil {
		return true
	}

	return r.Think == "" && r.Action == "" && r.Question == "" && r.Result == ""
}

func (r *RunStep) MergeWith(src *RunStep) {
	if src == nil {
		return
	}
	if src.Action != "" {
		r.Action = src.Action
	}
	if src.Question != "" {
		r.Question = src.Question
	}
	if src.Think != "" {
		r.Think = src.Think
	}
	if src.Result != "" {
		r.Result = src.Result
	}
	if src.State > RunState_Idle {
		r.State = src.State
	}
	if src.StepType > StepType_None {
		r.StepType = src.StepType
	}
}

const (
	defaultPromptReplaceContext = "{{context}}"
	defaultPromptReplaceTools   = "{{tools}}"
	defaultPromptReplaceSession = "{{session}}"
	defaultPromptReplaceAgents  = "{{agents}}"
)

// RunState 表示当前状态
type RunState int

const (
	RunState_Idle RunState = iota
	RunState_Running
	RunState_Succeed
	RunState_Failed
	RunState_Error
)

// String 返回状态的字符串表示
func (s RunState) String() string {
	switch s {
	case RunState_Idle:
		return "idle"
	case RunState_Running:
		return "running"
	case RunState_Succeed:
		return "succeed"
	case RunState_Failed:
		return "failed"
	case RunState_Error:
		return "error"
	default:
		return "unknown"
	}
}

// StepType 表示步骤类别
type StepType int

const (
	StepType_None StepType = iota
	StepType_Start
	StepType_End
	StepType_Tool
	StepType_Agent
)

// String 返回状态的字符串表示
func (s StepType) String() string {
	switch s {
	case StepType_Start:
		return "START"
	case StepType_End:
		return "END"
	case StepType_Tool:
		return "TOOLCALL"
	case StepType_Agent:
		return "AGENTCALL"
	default:
		return "UNKNOWN"
	}
}

const prettyCommonTpl = `
**%s：**
'''
%s
'''
`

const prettyStepHasThinkTpl = `
**第%d步➡️：**
- **思考🤖‍：** 
'''
%s
'''
- **执行🏃‍：** 
'''
%s => %s(%s)
'''
- **结果✅：** 
'''
%s
'''
`

const prettyStepTpl = `
**第%d步➡️：**
- **执行🏃‍：** 
'''
%s => %s(%s)
'''
- **结果✅：** 
'''
%s
'''
`

const prettyClaimTpl = `
----------
'''
ℹ️ %s
'''
`

func (opts *RunOptions) UpdateSystemPrompt(content string) string {
	if opts != nil && opts.Context != nil {
		contentBS, _ := json.Marshal(opts.Context)
		content = strings.Replace(content, defaultPromptReplaceContext, string(contentBS), -1)
	}
	if opts.SessionData != nil && len(opts.SessionData) > 0 {
		toolsBS, _ := json.Marshal(opts.SessionData)
		content = strings.Replace(content, defaultPromptReplaceSession, string(toolsBS), -1)
	}
	if opts.Tools != nil && len(opts.Tools) > 0 {
		toolsBS, _ := json.Marshal(opts.Tools)
		content = strings.Replace(content, defaultPromptReplaceTools, string(toolsBS), -1)
	}
	if opts.Agents != nil && len(opts.Agents) > 0 {
		agentsBS, _ := json.Marshal(opts.Agents)
		content = strings.Replace(content, defaultPromptReplaceAgents, string(agentsBS), -1)
	}
	return content
}

func (opts *RunOptions) CheckStepQuit() bool {
	opts.lock.RLock()
	defer opts.lock.RUnlock()

	if opts.steps == nil {
		return false
	}

	// 超过最大步数跳出
	return len(opts.steps) > opts.RuntimeCfg.MaxStepQuit
}

func (opts *RunOptions) AddStep(src *RunStep) {
	opts.lock.Lock()
	defer opts.lock.Unlock()

	if src.Action == AgentCallFuncName {
		src.StepType = StepType_Agent
	}

	tmp1 := &RunStep{}
	tmp2 := &RunStep{}
	if src.Result != "" {
		json.Unmarshal([]byte(src.Result), tmp1) // 兼容AgentCall模式
	}
	if src.Question != "" {
		json.Unmarshal([]byte(src.Question), tmp2) // 兼容AgentCall模式
	}

	if !tmp1.IsEmpty() {
		src.MergeWith(tmp1)
	} else if !tmp2.IsEmpty() {
		src.MergeWith(tmp2)
	}

	src.EndTime = time.Now()
	opts.steps = append(opts.steps, src)
}

func (opts *RunOptions) RenderFinalAnswer() string {
	opts.lock.RLock()
	defer opts.lock.RUnlock()

	output := ""
	for idx, step := range opts.steps {
		switch step.StepType {
		case StepType_Start:
			// output += fmt.Sprintf(prettyCommonTpl, "用户问题🤔", step.Question)
		case StepType_End:
			if HasMarkdownSyntax(step.Result) {
				output += "**最终结果📤:**\n" + strings.Trim(step.Result, "\n")
			} else {
				// 最终结果无格式输出才替换
				output += fmt.Sprintf(prettyCommonTpl, "最终结果📤", strings.Trim(step.Result, "\n"))
			}
		default:
			if opts.RuntimeCfg.Debug {
				if step.Think != "" {
					output += fmt.Sprintf(prettyStepHasThinkTpl, idx,
						strings.Trim(step.Think, "\n"),
						step.StepType.String(),
						strings.Trim(step.Action, "\n"),
						strings.Trim(step.Question, "\n"),
						strings.Trim(step.Result, "\n"),
					)
				} else {
					output += fmt.Sprintf(prettyStepTpl, idx,
						step.StepType.String(),
						strings.Trim(step.Action, "\n"),
						strings.Trim(step.Question, "\n"),
						strings.Trim(step.Result, "\n"),
					)
				}
			}
		}

	}

	if opts.RuntimeCfg.Claim != "" && output != "" {
		output += fmt.Sprintf(prettyClaimTpl, strings.Trim(opts.RuntimeCfg.Claim, "\n"))
	}

	return strings.TrimLeft(strings.Replace(output, "'''", "```", -1), "\n")
}

// RunOptionFunc 运行时选项
type RunOptionFunc func(*RunOptions)

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

func WithSystemPrompt(systemPrompt string) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.RuntimeCfg.SystemPrompt = systemPrompt
	}
}

func WithContext(context interface{}) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Context = context
	}
}

func WithSessionID(sessionID string) RunOptionFunc {
	return func(opts *RunOptions) {
		if opts.Session != nil {
			opts.Session.SessionID = sessionID
		}
	}
}

func WithSessionData(sessionData map[string]interface{}) RunOptionFunc {
	return func(opts *RunOptions) {
		if opts.Session != nil {
			opts.Session.MergeSessionData(sessionData)
		}
	}
}

func WithAgents(agents []string) RunOptionFunc {
	return func(opts *RunOptions) {
		if opts.Agents == nil {
			opts.Agents = make([]BriefInfo, 0)
		}

		ags := GetAgentHub().GetAgentList(agents...)
		for _, ag := range ags {
			opts.Agents = append(opts.Agents, ag.GetBriefInfo())
		}
	}
}
