package ssestream

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
)

const streamTpl = "event: %s\ndata: %s\r\n\r\n"

type Encoder interface {
	Close() error
	Event() <-chan Event
	Encode(src interface{}) bool
	Send(event Event) (err error)
}

func NewEncoder(writer io.WriteCloser) Encoder {
	if writer == nil {
		return nil
	}

	encoder := &eventStreamEncoder{
		eventCh: make(chan Event, 100),
		wc:      writer,
	}
	encoder.flusher, _ = writer.(http.Flusher)
	return encoder
}

type eventStreamEncoder struct {
	eventCh chan Event
	wc      io.WriteCloser
	flusher http.Flusher
}

func (s *eventStreamEncoder) Send(event Event) (err error) {
	stream := fmt.Sprintf(streamTpl, event.Type, string(event.Data))
	_, err = s.wc.Write([]byte(stream))
	if err == nil {
		if s.flusher != nil {
			s.flusher.Flush()
		}
	}
	return
}

func (s *eventStreamEncoder) Close() error {
	return s.wc.Close()
}

func (s *eventStreamEncoder) Event() <-chan Event {
	return s.eventCh
}

func (s *eventStreamEncoder) Encode(src interface{}) bool {
	event := Event{
		Type: "response.append",
		Data: []byte{},
	}
	event.Data, _ = json.Marshal(src)
	if ep := gjson.GetBytes(event.Data, "error"); ep.Exists() {
		event.Type = "response.error"
	}

	select {
	case s.eventCh <- event:
		return true
	default:
		// full
		log.Printf("eventStreamEncoder.Encode: channel full not push => %s", string(event.Data))
	}
	return false
}

type StreamWriter[T any] struct {
	encoder Encoder
	closed  bool
	err     error
	ctx     context.Context
}

func NewStreamWriter[T any](encoder Encoder, ctx context.Context) *StreamWriter[T] {
	ret := &StreamWriter[T]{
		ctx:     ctx,
		encoder: encoder,
	}

	go ret.runloop()
	return ret
}

func (s *StreamWriter[T]) runloop() {
	for {
		if s.closed {
			return
		}

		select {
		case event := <-s.encoder.Event():
			if s.err = s.encoder.Send(event); s.err != nil {
				s.Close()
				return
			}
		case <-s.ctx.Done():
			s.Close()
			return
		}
	}
}

func (s *StreamWriter[T]) Append(t *T) bool {
	if s.closed {
		return false
	}
	return s.encoder.Encode(t)
}

func (s *StreamWriter[T]) Err() error {
	return s.err
}

func (s *StreamWriter[T]) Close() error {
	if s.closed {
		return nil
	}

	if s.err == nil {
		// 正常关闭，先发送完[DONE]
		s.encoder.Send(Event{
			Type: "response.done",
			Data: []byte("[DONE]"),
		})
	}

	s.closed = true
	return s.encoder.Close()
}
