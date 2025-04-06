/*
@Project: aihub
@Module: core
@File : tool.go
*/
package aihub

import (
	"github.com/mvptianyu/aihub/jsonschema"
)

const (
	ToolTypeFunction = "function"
)

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  *jsonschema.Definition `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Strict      bool                   `json:"strict,omitempty" yaml:"strict,omitempty"`
}

// IToolInput 工具入参格式定义
type IToolInput interface {
	GetRawInput() string
	SetRawInput(str string)
}

type ToolInputBase struct {
	input string `json:"-"`
}

func (t *ToolInputBase) GetRawInput() string {
	return t.input
}

func (t *ToolInputBase) SetRawInput(str string) {
	t.input = str
}
