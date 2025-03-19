package core

import "github.com/mvptianyu/aihub/types"

// Memory interface for different memory implementations
type MemoryStorer interface {
	Push(m ...*types.Message)
	Peek() *types.Message
	Dump() []*types.Message

	Clear()
}
