package ssestream

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"strings"
)

type Decoder interface {
	Event() Event
	Next() bool
	Close() error
	Err() error
}

func NewDecoder(reader io.ReadCloser) Decoder {
	if reader == nil {
		return nil
	}
	scanner := bufio.NewScanner(reader)
	decoder := &eventStreamDecoder{rc: reader, scn: scanner}
	return decoder
}

type Event struct {
	Type string
	Data []byte
}

// A base implementation of a Decoder for text/event-stream.
type eventStreamDecoder struct {
	evt Event
	rc  io.ReadCloser
	scn *bufio.Scanner
	err error
}

func (s *eventStreamDecoder) Next() bool {
	if s.err != nil {
		return false
	}

	event := ""
	data := bytes.NewBuffer(nil)

	for s.scn.Scan() {
		txt := s.scn.Bytes()

		// Dispatch event on an empty line
		if len(txt) == 0 {
			s.evt = Event{
				Type: event,
				Data: data.Bytes(),
			}
			return true
		}

		// Split a string like "event: bar" into name="event" and value=" bar".
		name, value, _ := bytes.Cut(txt, []byte(":"))

		// Consume an optional space after the colon if it exists.
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}

		switch string(name) {
		case "":
			// An empty line in the for ": something" is a comment and should be ignored.
			continue
		case "event":
			event = string(value)
		case "data":
			_, s.err = data.Write(value)
			if s.err != nil {
				break
			}
			_, s.err = data.WriteRune('\n')
			if s.err != nil {
				break
			}
		}
	}

	if s.scn.Err() != nil {
		s.err = s.scn.Err()
	}

	return false
}

func (s *eventStreamDecoder) Event() Event {
	return s.evt
}

func (s *eventStreamDecoder) Close() error {
	return s.rc.Close()
}

func (s *eventStreamDecoder) Err() error {
	return s.err
}

type StreamReader[T any] struct {
	decoder Decoder
	cur     T
	err     error
	done    bool
}

func NewStreamReader[T any](decoder Decoder, err error) *StreamReader[T] {
	return &StreamReader[T]{
		decoder: decoder,
		err:     err,
	}
}

// Next returns false if the stream has ended or an error occurred.
// Call StreamReader.Current() to get the current value.
// Call StreamReader.Err() to get the error.
//
//		for stream.Next() {
//			data := stream.Current()
//		}
//
//	 	if stream.Err() != nil {
//			...
//	 	}
func (s *StreamReader[T]) Next() bool {
	if s.err != nil {
		return false
	}

	for s.decoder.Next() {
		if s.done {
			continue
		}

		if bytes.HasPrefix(s.decoder.Event().Data, []byte("[DONE]")) {
			// In this case we don't break because we still want to iterate through the full stream.
			s.done = true
			continue
		}

		if s.decoder.Event().Type == "" || strings.HasPrefix(s.decoder.Event().Type, "response.") {
			ep := gjson.GetBytes(s.decoder.Event().Data, "error")
			if ep.Exists() {
				s.err = fmt.Errorf("received error while streaming: %s", ep.String())
				return false
			}
			s.err = json.Unmarshal(s.decoder.Event().Data, &s.cur)
			if s.err != nil {
				return false
			}
			return true
		} else {
			ep := gjson.GetBytes(s.decoder.Event().Data, "error")
			if ep.Exists() {
				s.err = fmt.Errorf("received error while streaming: %s", ep.String())
				return false
			}
			event := s.decoder.Event().Type
			data := s.decoder.Event().Data
			s.err = json.Unmarshal([]byte(fmt.Sprintf(`{ "event": %q, "data": %s }`, event, data)), &s.cur)
			if s.err != nil {
				return false
			}
			return true
		}
	}

	// decoder.Next() may be false because of an error
	s.err = s.decoder.Err()

	return false
}

func (s *StreamReader[T]) Current() T {
	return s.cur
}

func (s *StreamReader[T]) Err() error {
	return s.err
}

func (s *StreamReader[T]) Close() error {
	if s.decoder == nil {
		// already closed
		return nil
	}
	return s.decoder.Close()
}
