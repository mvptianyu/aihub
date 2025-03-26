/*
@Project: aihub
@Module: core
@File : tool.go
*/
package core

import "context"

type WrapToolFunc func(ctx context.Context, in []byte) (out []byte, err error)

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

type Tool struct {
	Type     ToolType     `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                  `json:"name" yaml:"name"`
	Description string                  `json:"description,omitempty" yaml:"name,omitempty"`
	Parameters  *ToolFunctionParameters `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Strict      bool                    `json:"strict,omitempty" yaml:"strict,omitempty"`

	wrapToolFunc WrapToolFunc `json:"-" yaml:"-"` // 执行入口
	jsonSchema   []byte       `json:"-" yaml:"-"` // 参数信息
}

type ToolFunctionParametersType string

const (
	ToolFunctionParametersTypeText   ToolFunctionParametersType = "text"
	ToolFunctionParametersTypeObject ToolFunctionParametersType = "object"
)

type ToolFunctionParameters struct {
	Type       ToolFunctionParametersType `json:"type" yaml:"type"`
	Properties map[string]interface{}     `json:"properties" yaml:"properties"`
	Required   []string                   `json:"required" yaml:"required"`
}
