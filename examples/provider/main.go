package main

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/core"
	"time"
)

func initProvider() core.IProvider {
	return aihub.NewProvider(&core.ProviderConfig{
		Name:    "openai",
		Model:   "gpt-3.5-turbo",
		BaseURL: "https://api.openai.com",
	})
}

func main() {
	ctx := context.Background()

	myProvider := initProvider()

	req := &core.CreateChatCompletionReq{
		Messages: []*core.Message{
			{
				Content: "请你评价以下AI Agent技术的现状，不超过200字",
				Role:    core.MessageRoleUser,
			},
		},
		Model: "gpt-3.5-turbo",
	}

	CreateChatCompletion(ctx, myProvider, req)
	//CreateChatCompletionStream(ctx, myProvider, req)
}

func CreateChatCompletion(ctx context.Context, myProvider core.IProvider, req *core.CreateChatCompletionReq) {
	rsp, err := myProvider.CreateChatCompletion(ctx, req)

	fmt.Println(err)
	fmt.Println("=======================")
	fmt.Println(rsp.Choices[0])
}

func CreateChatCompletionStream(ctx context.Context, myProvider core.IProvider, req *core.CreateChatCompletionReq) {
	stream := myProvider.CreateChatCompletionStream(ctx, req)

	for stream.Next() {
		data := stream.Current()
		fmt.Printf(data.Choices[0].Delta.Content)
		time.Sleep(10 * time.Millisecond)
		// fmt.Println(data)
	}

	fmt.Println("\n======[Done]=======")
}
