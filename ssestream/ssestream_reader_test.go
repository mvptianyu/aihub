/*
@Project: mvptianyu
@Module: aihub
@File : ssestream_test.go
*/
package ssestream

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"
)

type Response struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

func getStreamReader(ctx context.Context) *StreamReader[string] {
	var err error
	r, w := io.Pipe()
	writer := NewStreamWriter[string](NewEncoder(w), ctx)
	reader := NewStreamReader[string](NewDecoder(r), err)

	go func() {
		for i := 0; i < 20; i++ {
			rsp := &Response{}
			rsp.Content = fmt.Sprintf("本次内容：%d", i)
			if i == 18 {
				rsp.Error = "xxxxx error"
			}

			writer.Append(&rsp.Content)
			time.Sleep(300 * time.Millisecond)
		}
		writer.Close()
	}()

	return reader
}

func TestStream(t *testing.T) {
	stream := getStreamReader(context.Background())

	for stream.Next() {
		data := stream.Current()
		fmt.Printf(data)
	}
	fmt.Println(stream.Err())
	fmt.Println("\n======[Done]=======")
}
