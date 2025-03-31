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

type ToolFunction struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  *jsonschema.Definition `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Strict      bool                   `json:"strict,omitempty" yaml:"strict,omitempty"`
}

// IToolInput 工具入参格式定义
type IToolInput interface {
	GetRawInput() string
	GetRawCallID() string
	GetRawSession() map[string]interface{}
	SetRawInput(str string)
	SetRawCallID(str string)
	SetRawSession(session map[string]interface{})
}

type ToolInputBase struct {
	input   string                 `json:"-"`
	callID  string                 `json:"-"`
	session map[string]interface{} `json:"-"`
}

func (t *ToolInputBase) GetRawInput() string {
	return t.input
}

func (t *ToolInputBase) GetRawCallID() string {
	return t.callID
}

func (t *ToolInputBase) GetRawSession() map[string]interface{} {
	return t.session
}

func (t *ToolInputBase) SetRawInput(str string) {
	t.input = str
}

func (t *ToolInputBase) SetRawCallID(str string) {
	t.callID = str
}

func (t *ToolInputBase) SetRawSession(session map[string]interface{}) {
	t.session = session
}
