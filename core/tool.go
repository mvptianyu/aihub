/*
@Project: aihub
@Module: core
@File : tool.go
*/
package core

import (
	"context"
	"github.com/mvptianyu/aihub/jsonschema"
)

type ToolFuncRouter func(ctx context.Context, name string, in string) (out interface{}, err error)

const (
	ToolTypeFunction         = "function"
	ToolArgumentsRawInputKey = "_INPUT_"
)

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolSummary struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type ToolFunction struct {
	ToolSummary `yaml:",inline"`
	Parameters  *jsonschema.Definition `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Strict      bool                   `json:"strict,omitempty" yaml:"strict,omitempty"`
}

// IToolInput 工具入参格式定义
type IToolInput interface {
	GetRawInput() string
	GetRawCallID() string
	GetRawFuncName() string
	SetRawInput(str string)
	SetRawCallID(str string)
	SetRawFuncName(str string)
}

type ToolInputBase struct {
	input    string `json:"-"`
	callID   string `json:"-"`
	funcName string `json:"-"`
}

func (t *ToolInputBase) GetRawInput() string {
	return t.input
}

func (t *ToolInputBase) GetRawCallID() string {
	return t.callID
}
func (t *ToolInputBase) GetRawFuncName() string {
	return t.funcName
}

func (t *ToolInputBase) SetRawInput(str string) {
	t.input = str
}

func (t *ToolInputBase) SetRawCallID(str string) {
	t.callID = str
}

func (t *ToolInputBase) SetRawFuncName(str string) {
	t.funcName = str
}
