/*
@Project: aihub
@Module: core
@File : debuger.go
*/
package core

import (
	"fmt"
	"strings"
	"sync"
)

/*
React模式Prompt模版：

Question: the input question you must answer
Thought: you should always think about what to do
Action: the action to take, should be one of [{tool_names}]
Action Input: the input to the action
Observation: the result of the action
... (this Thought/Action/Action Input/Observation can be repeated zero or more times)
Thought: I now know the final answer
Final Answer: the final answer to the original input question
*/

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

type Recorder struct {
	Question    string
	Steps       []*RecordStep
	FinalAnswer string

	lock sync.RWMutex
}

type RecordStep struct {
	Action      string
	Observation string
}

func NewRecorder() *Recorder {
	return &Recorder{
		Steps: make([]*RecordStep, 0),
	}
}

func (r *Recorder) AddStep(toolCalls []*MessageToolCall, toolMsgs []*Message) {
	r.lock.Lock()
	defer r.lock.Unlock()

	action := ""
	observation := ""

	for _, toolCall := range toolCalls {
		action += fmt.Sprintf("%s => %s( %s )\n", toolCall.Id, toolCall.Function.Name, toolCall.Function.Arguments)
	}

	for _, toolMsg := range toolMsgs {
		observation += fmt.Sprintf("%s => %s\n", toolMsg.ToolCallID, toolMsg.Content)
	}

	r.Steps = append(r.Steps, &RecordStep{
		Action:      strings.TrimRight(action, "\n"),
		Observation: strings.TrimRight(observation, "\n"),
	})
}

func (r *Recorder) SetQuestion(question string) {
	r.Question = question
}

func (r *Recorder) SetFinal(final string) {
	r.FinalAnswer = final
}

func (r *Recorder) PrettyPrint() string {
	output := fmt.Sprintf(prettyCommonTpl, "用户问题🤔", r.Question)
	for idx, step := range r.Steps {
		output += fmt.Sprintf(prettyStepTpl, idx+1, step.Action, step.Observation)
	}
	if HasMarkdownSyntax(r.FinalAnswer) {
		output += r.FinalAnswer
	} else {
		// 最终结果无格式输出才替换
		output += fmt.Sprintf(prettyCommonTpl, "最终结果📤", r.FinalAnswer)
	}
	return strings.TrimLeft(strings.Replace(output, "'''", "```", -1), "\n")
}
