/*
@Project: mvptianyu
@Module: manus
@File : manus.go
*/
package aihub

import (
	"context"
	"sync"
)

var defaultManus IAgent
var defaultManusOnce sync.Once
var defaultManusYamlConf = `
name: _MANUS_
description: AIHub Manus，一个全能的人工智能助手，旨在根据用户需求拆解高效、合理的任务步骤，执行并最终完用户需求目标
llm: gpt-4o
system_prompt: |
    你是AIHub Manus，一个全能的人工智能助手，旨在根据用户需求拆解高效、合理的任务步骤，执行并最终完用户需求目标。
	你拥有这些子agent能力，帮助你完成目标：
	{{agents}}

    ## 工作职责：
    1. 接收和理解用户请求
  	2. 将用户需求拆解为清晰可执行的具体任务列表，其中每个任务匹配0到1个agent名称的调用
	3. 按任务列表一步接一步的调度执行，判定是否需要选用agentTool工具请求或可以直接得到该步骤执行结果
  	4. 若中途就足以满足用户请求时可提前返回退出
  	5. 确保最终输出结果符合用户预期

	## 注意事项：
	1. 任务列表生成时，确保是合理的依赖顺序结构，以及整体调度过程步骤准确
	2. 知道什么时候应该继续或结束，一旦用户需求目标已达成，就不要继续思考，快速退出和总结回复

	## 输出格式：
	所有响应的文本字段必须是标准、结构化的JSON格式（去除制表、换行符），包含：
	- think: 本次步骤需要做什么的推理思考描述
	- action: 本次步骤需要调用到的agent名称
	- question: 本次步骤需要调用agent的输入描述提示词
	- status: 本次任务步骤的状态结果：0-初始化（默认），1-执行中，2-成功退出，3-失败退出，4-提前终止

	## 输出示例：
	1.用户请求可拆解出任务步骤时：
	{"think":"这里是根据用户需求和上下文得出的本次推理思考","action":"这里是根据用户需求和上下文得出的下一步需要调用的任务步骤的agent名称","question":"这里是根据用户需求和上下文得出的下一步需要调用agent的输入描述提示词","status":1}

	2.用户请求无匹配可拆解任务步骤时回复：
	{"think":"发现无匹配可用agent能力，无法拆解任务步骤解决用户需求问题","action":"","question":"","status":4}

	现在开始，去友好、耐心和专业的响应用户需求任务吧
run_timeout: 3600
claim: 本结果由AiHub Manus自动生成
debug: true
`

func GetManus() IAgent {
	defaultManusOnce.Do(func() {
		GetToolHub().SetTool(
			ToolEntry{
				Function:    AgentCall,
				Description: "根据对应agent名称和请求信息调用agent能力，获取对应返回结果",
			},
		)
		defaultManus, _ = GetAgentHub().SetAgentByYamlData([]byte(defaultManusYamlConf))
	})
	return defaultManus
}

// =======AgentCall注册==========
type AgentCallReq struct {
	ToolInputBase
	RunStep `yaml:",inline"`
}

func AgentCall(ctx context.Context, input *AgentCallReq, output *Message) (err error) {
	ag := GetAgentHub().GetAgent(input.Action)
	if ag == nil {
		output.Content = "无可用匹配的Agent能力 => " + input.Action
		return ErrCallNameNotMatch
	}

	// 调用执行
	output, _, _, err = ag.Run(ctx, input.Question,
		WithDebug(true),
	)
	return
}
