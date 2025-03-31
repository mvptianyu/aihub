package core

import (
	"sync"
	"time"
)

type memory struct {
	sysMessage *Message
	messages   map[string][]*Message // sessionid => msgList
	limit      int
	timeout    int64
	lock       sync.RWMutex
}

func NewMemory(limit int, timeout int64) *memory {
	ret := &memory{
		messages: make(map[string][]*Message, 0),
		limit:    limit,
		timeout:  timeout,
	}

	go ret.cronClean()
	return ret
}

func (h *memory) cronClean() {
	h.lock.Lock()
	defer h.lock.Unlock()

	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		// 过期判定和清理
		now := time.Now().Unix()
		for sessionId, messageList := range h.messages {
			newList := make([]*Message, 0)
			for _, message := range messageList {
				if now-message.CreateTime < h.timeout {
					newList = append(newList, message)
				}
			}
			if len(newList) > 0 {
				h.messages[sessionId] = newList
			} else {
				delete(h.messages, sessionId)
			}
		}
	}
}

func (h *memory) GetSystemMsg() *Message {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.sysMessage.Copy()
}

func (h *memory) SetSystemMsg(msg *Message) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.sysMessage = msg
}

func (h *memory) Push(opts *RunOptions, msg ...*Message) {
	toAddCnt := len(msg)

	h.lock.Lock()
	defer h.lock.Unlock()

	for _, singleMsg := range msg {
		if singleMsg.CreateTime == 0 {
			singleMsg.CreateTime = opts.CreateTime
		}
	}

	if _, ok := h.messages[opts.SessionID]; !ok {
		h.messages[opts.SessionID] = make([]*Message, 0)
	}

	oralLen := len(h.messages[opts.SessionID])
	if oralLen+toAddCnt > h.limit {
		shiftIdx := toAddCnt - (h.limit - oralLen)
		h.messages[opts.SessionID] = h.messages[opts.SessionID][shiftIdx:]
	}
	h.messages[opts.SessionID] = append(h.messages[opts.SessionID], msg...)
}

func (h *memory) GetLatest(opts *RunOptions) []*Message {
	h.lock.RLock()
	defer h.lock.RUnlock()

	target, ok := h.messages[opts.SessionID]
	if !ok {
		return nil
	}

	idx := 0
	if opts.RuntimeCfg.MaxUseHistory > 0 && opts.RuntimeCfg.MaxUseHistory < len(target) {
		idx = len(target) - opts.RuntimeCfg.MaxUseHistory
	}

	tmp := make([]*Message, 0)
	if opts.RuntimeCfg.SystemPrompt != "" && h.sysMessage != nil {
		tmp = append(tmp, h.sysMessage.Copy())
	}
	return append(tmp, target[idx:]...)
}

func (h *memory) Clear(opts *RunOptions) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.messages[opts.SessionID]; ok {
		delete(h.messages, opts.SessionID)
	} else {
		h.messages = make(map[string][]*Message, 0) // 删除所有
	}
}
