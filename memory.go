package aihub

import (
	"sync"
	"time"
)

type memory struct {
	messages map[string][]*Message // sessionid => msgList
	limit    int
	timeout  int64
	lock     sync.RWMutex
}

func newMemory(cfg *AgentRuntimeCfg) IMemory {
	ret := &memory{
		messages: make(map[string][]*Message),
		limit:    cfg.MaxStoreMemory,
		timeout:  cfg.MemoryTimeout,
	}

	go ret.cronClean()
	return ret
}

func (h *memory) cronClean() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		// 过期判定和清理
		now := time.Now().Unix()

		h.lock.Lock()
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
		h.lock.Unlock()
	}
}

func (h *memory) Push(opts *RunOptions, msg ...*Message) {
	toAddCnt := len(msg)
	now := time.Now().Unix()

	h.lock.Lock()
	defer h.lock.Unlock()

	sessionId := opts.GetSessionID()
	for _, singleMsg := range msg {
		if singleMsg.CreateTime == 0 {
			singleMsg.CreateTime = now
		}
		if singleMsg.SessionID == "" {
			singleMsg.SessionID = sessionId
		}
	}

	if _, ok := h.messages[sessionId]; !ok {
		h.messages[sessionId] = make([]*Message, 0)
	}

	oralLen := len(h.messages[sessionId])
	if oralLen+toAddCnt > h.limit {
		shiftIdx := toAddCnt - (h.limit - oralLen)
		h.messages[sessionId] = h.messages[sessionId][shiftIdx:]
	}
	h.messages[sessionId] = append(h.messages[sessionId], msg...)
}

func (h *memory) GetLatest(opts *RunOptions) []*Message {
	h.lock.RLock()
	target, ok := h.messages[opts.GetSessionID()]
	h.lock.RUnlock()
	if !ok {
		return []*Message{}
	}

	idx := 0
	if opts.RuntimeCfg.MaxUseMemory > 0 && opts.RuntimeCfg.MaxUseMemory < len(target) {
		idx = len(target) - opts.RuntimeCfg.MaxUseMemory
	}
	return target[idx:]
}

func (h *memory) Clear(opts *RunOptions) {
	h.lock.Lock()
	defer h.lock.Unlock()

	sessionId := opts.GetSessionID()
	if _, ok := h.messages[sessionId]; ok {
		delete(h.messages, sessionId)
	} else {
		h.messages = make(map[string][]*Message) // 删除所有
	}
}
