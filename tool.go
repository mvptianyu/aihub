package aihub

import (
	"github.com/mvptianyu/aihub/jsonschema"
)

const (
	ToolArgumentsRawInputKey   = "INPUT_"
	ToolArgumentsRawSessionKey = "SESSION_"
)

const (
	ToolTypeFunction = "function"
)

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	BriefInfo
	Parameters *jsonschema.Definition `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Strict     bool                   `json:"strict,omitempty" yaml:"strict,omitempty"`
}

// IToolInput 工具入参格式定义
type IToolInput interface {
	GetRawInput() string
	SetRawInput(str string)
	GetRawSession() string
	SetRawSession(str string)
}

type ToolInputBase struct {
	input   string `json:"-"`
	Session string `json:"SESSION_" description:"记录session的key名,默认为空,无需设置" required:"false"`
}

func (t *ToolInputBase) GetRawInput() string {
	return t.input
}

func (t *ToolInputBase) SetRawInput(str string) {
	t.input = str
}
func (t *ToolInputBase) GetRawSession() string {
	return t.Session
}

func (t *ToolInputBase) SetRawSession(str string) {
	t.Session = str
}
