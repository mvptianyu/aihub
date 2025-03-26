package core

import "sync"

type history struct {
	sysMessage *Message
	messages   []*Message
	limit      int
	lock       sync.RWMutex
}

func NewHistory(limit int) *history {
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	return &history{
		messages: make([]*Message, 0),
		limit:    limit,
	}
}

func (h *history) GetSystemMsg() *Message {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.sysMessage
}

func (h *history) SetSystemMsg(msg *Message) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.sysMessage = msg
}

func (h *history) Push(msg ...*Message) {
	toAddCnt := len(msg)

	h.lock.Lock()
	defer h.lock.Unlock()

	if len(h.messages)+toAddCnt >= h.limit {
		h.messages = h.messages[toAddCnt:]
	}
	h.messages = append(h.messages, msg...)
}

func (h *history) Pop() *Message {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if len(h.messages) == 0 {
		return nil
	}

	return h.messages[len(h.messages)-1]
}

func (h *history) GetAll(length int, needSystem bool) []*Message {
	h.lock.RLock()
	defer h.lock.RUnlock()

	idx := 0
	if length > 0 || length < len(h.messages) {
		idx = len(h.messages) - length
	}

	tmp := make([]*Message, 0)
	if needSystem && h.sysMessage != nil {
		tmp = append(tmp, h.sysMessage)
	}
	return append(tmp, h.messages[idx:]...)
}

func (h *history) Clear() {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.messages = make([]*Message, 0)
}
