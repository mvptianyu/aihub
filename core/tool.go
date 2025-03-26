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
	ToolFunctionDefaultParam = "_INPUT_"
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
