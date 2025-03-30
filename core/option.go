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
)

type RunOptions struct {
	Recorder // 上下文记录器

	StopWords string // 结束退出词
	Debug     bool   // debug标志，开启则输出具体工具调用过程信息
	Claim     string // 宣称文案，例如：本次返回由xxx提供
	Context   string // 上下文提示词相关，例如在systemprompt中插入/替换该内容
	CurStep   int

	toolPrompts []*ToolFunction
}

func NewRunOptions(a *Agent) *RunOptions {
	return &RunOptions{
		toolPrompts: a.cfg.Tools,
	}
}

const (
	defaultPromptReplaceContext = "{{context}}"
	defaultPromptReplaceTools   = "{{toolMethods}}"
)

func (opts *RunOptions) FixMessageContent(role MessageRoleType, content string) string {
	switch role {
	case MessageRoleSystem:
		if opts != nil && opts.Context != "" {
			contentBS, _ := json.Marshal(opts.Context)
			content = strings.Replace(content, defaultPromptReplaceContext, string(contentBS), -1)
		}
		if opts.toolPrompts != nil && len(opts.toolPrompts) > 0 {
			toolsBS, _ := json.Marshal(opts.toolPrompts)
			content = strings.Replace(content, defaultPromptReplaceTools, string(toolsBS), -1)
		}
	case MessageRoleAssistant:
		if opts.Debug && content != "" {
			content = opts.PrettyPrint()
		}

		if opts.Claim != "" && content != "" {
			content += fmt.Sprintf("\n```ℹ️ %s```", opts.Claim)
		}
	default:
	}
	return content
}

// ------------

type RunOptionFunc func(*RunOptions)

func WithStopWords(StopWords string) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.StopWords = StopWords
	}
}

func WithDebug(Debug bool) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Debug = Debug
	}
}

func WithClaim(Claim string) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Claim = Claim
	}
}

func WithContext(Context string) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Context = Context
	}
}
