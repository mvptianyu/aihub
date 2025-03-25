package core

type history struct {
	messages []*Message
	limit    int
}

func NewHistory(limit int) *history {
	if limit <= 0 {
		limit = 10 // 默认最近10个
	}

	return &history{
		messages: make([]*Message, limit),
		limit:    limit,
	}

}

func (h *history) Push(msg ...*Message) {
	h.messages = append(h.messages, msg...)
}

func (h *history) Peek() *Message {
	if len(h.messages) == 0 {
		return nil
	}

	return h.messages[len(h.messages)-1]
}

func (h *history) Dump() []*Message {
	return h.messages
}

func (h *history) Clear() {
	h.messages = []*Message{}
}
