package core

import "github.com/mvptianyu/aihub/types"

// IMemoryStorer 历史会话记忆
type IMemoryStorer interface {
	Push(m ...*types.Message)
	Peek() *types.Message
	Dump() []*types.Message
	Clear()
}
