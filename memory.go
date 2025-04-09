package aihub

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

func (h *memory) GetSystemMsg() *Message {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if h.sysMessage == nil {
		return nil
	}

	return h.sysMessage.Copy()
}

func (h *memory) SetSystemMsg(msg *Message) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.sysMessage = msg
}

func (h *memory) Push(opts *RunOptions, msg ...*Message) {
	toAddCnt := len(msg)
	now := time.Now().Unix()

	h.lock.Lock()
	defer h.lock.Unlock()

	for _, singleMsg := range msg {
		if singleMsg.CreateTime == 0 {
			singleMsg.CreateTime = now
		}
		if singleMsg.SessionID == "" {
			singleMsg.SessionID = opts.SessionID
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
	target, ok := h.messages[opts.SessionID]
	h.lock.RUnlock()
	if !ok {
		return nil
	}

	idx := 0
	if opts.RuntimeCfg.MaxUseMemory > 0 && opts.RuntimeCfg.MaxUseMemory < len(target) {
		idx = len(target) - opts.RuntimeCfg.MaxUseMemory
	}

	tmp := make([]*Message, 0)

	if sysMsg := h.GetSystemMsg(); sysMsg != nil {
		tmp = append(tmp, sysMsg)
	}
	return append(tmp, target[idx:]...)
}

func (h *memory) Clear(opts *RunOptions) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.messages[opts.SessionID]; ok {
		delete(h.messages, opts.SessionID)
	} else {
		h.messages = make(map[string][]*Message) // 删除所有
	}
}
