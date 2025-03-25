package core

import (
	"context"
)

type Middleware interface {
	Name() string

	Priority() uint

	PreProcess(ctx context.Context, m *Message) error

	PostProcess(ctx context.Context, m *Message) error
}
