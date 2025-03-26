/*
@Project: aihub
@Module: core
@File : provider_test.go
*/
package core

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_provider_CreateChatCompletion(t *testing.T) {
	p := NewProvider(&ProviderConfig{
		Name:      "openai",
		BaseURL:   "https://api.openai.com",
		Version:   "v1",
		APIKey:    "AAAAA",
		RateLimit: 3,
	})

	ctx := context.Background()
	question := "请你评价以下AI MCP技术的现状，不超过200字"

	req := &CreateChatCompletionReq{
		Model:    "gpt-4o",
		Messages: make([]*Message, 0),
	}
	message := &Message{
		Content: question,
	}
	message.Role = MessageRoleUser
	req.Messages = append(req.Messages, message)

	rsp, err := p.CreateChatCompletion(ctx, req)
	if err != nil {
		panic(err)
	}

	fmt.Printf(rsp.Choices[0].Message.Content)
	fmt.Println("\n======[Done]=======")
}

func Test_provider_CreateChatCompletionStream(t *testing.T) {
	p := NewProvider(&ProviderConfig{
		Name:      "openai",
		BaseURL:   "https://api.openai.com",
		Version:   "v1",
		APIKey:    "AAAAA",
		RateLimit: 3,
	})

	ctx := context.Background()
	question := "请你评价以下AI Agent技术的现状，不超过200字"

	req := &CreateChatCompletionReq{
		Model:    "gpt-4o",
		Messages: make([]*Message, 0),
	}
	message := &Message{
		Content: question,
		Role:    MessageRoleUser,
	}
	req.Messages = append(req.Messages, message)

	stream := p.CreateChatCompletionStream(ctx, req)

	for stream.Next() {
		data := stream.Current()
		fmt.Printf(data.Choices[0].Delta.Content)
		time.Sleep(20 * time.Millisecond)
		// fmt.Println(data)
	}

	fmt.Println("\n======[Done]=======")

	time.Sleep(5 * time.Second)
	if stream.Err() != nil {
		fmt.Println(stream.Err())
	}
}
