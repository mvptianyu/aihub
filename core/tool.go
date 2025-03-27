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
	ToolTypeFunction          = "function"
	ToolArgumentsRawInputKey  = "_INPUT_"
	ToolArgumentsRawCallIDKey = "_CALLID_"
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

func (t *ToolFunction) AutoFix() {
	if t.Name == "" {
		return
	}

	if t.Parameters == nil {
		t.Parameters = &jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				ToolArgumentsRawInputKey: {
					Type:        jsonschema.String,
					Description: "default input parameter",
				},
			},
			Required: []string{ToolArgumentsRawInputKey},
		}
	}
	if t.Description == "" {
		t.Description = t.Name
	}
}

// IToolInput 工具入参格式定义
type IToolInput interface {
	GetRawInput() string
	GetRawCallID() string
}

type ToolInputBase struct {
	input  string `json:"-"`
	callID string `json:"-"`
}

func (t ToolInputBase) GetRawInput() string {
	return t.input
}

func (t ToolInputBase) GetRawCallID() string {
	return t.callID
}
