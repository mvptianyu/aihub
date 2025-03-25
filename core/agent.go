/*
@Project: aihub
@Module: core
@File : agent.go
*/
package core

import (
	"context"
)

type Agent struct {
	cfg     *AgentConfig
	tools   map[string]Tool
	history *history
}

func NewAgent(cfg *AgentConfig) IAgent {
	return &Agent{
		cfg:     cfg,
		tools:   make(map[string]Tool),
		history: NewHistory(cfg.MaxChatHistory),
	}
}

func (a *Agent) Run(ctx context.Context, input string) (Message, error) {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) RunStream(ctx context.Context, input string) (<-chan Message, <-chan string, <-chan error) {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) GetTool(name string) (*Tool, bool) {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) RegisterTool(tool *Tool) error {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) RemoveTool(name string) error {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) ListTool() []*Tool {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) RegisterMiddleware(middle *Middleware) error {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) RemoveMiddleware(name string) error {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) GetMiddleware(name string) (Middleware, bool) {
	// TODO implement me
	panic("implement me")
}

func (a *Agent) ListMiddleware() []*Middleware {
	// TODO implement me
	panic("implement me")
}
