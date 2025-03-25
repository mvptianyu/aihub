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
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Parameters  *ToolFunctionParameters `json:"parameters,omitempty"`
	Strict      bool                    `json:"strict,omitempty"`

	wrapToolFunc WrapToolFunc `json:"-"` // 执行入口
	jsonSchema   []byte       `json:"-"` // 参数信息
}

type ToolFunctionParametersType string

const (
	ToolFunctionParametersTypeText   ToolFunctionParametersType = "text"
	ToolFunctionParametersTypeObject ToolFunctionParametersType = "object"
)

type ToolFunctionParameters struct {
	Type ToolFunctionParametersType `json:"type"`
}
