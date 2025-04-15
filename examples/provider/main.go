package main

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/examples/depency"
	"time"
)

func main() {
	depency.Init() // 初始化

	ctx := context.Background()
	myLLM, err := aihub.GetLLMHub().SetLLM(&aihub.LLMConfig{
		BriefInfo: aihub.BriefInfo{
			Name:        "gpt-3.5-turbo",
			Description: "openai's gpt-3.5-turbo LLM model api service",
		},
		Provider: "openai",
		BaseURL:  "https://api.openai.com",
	})

	if err != nil {
		panic(err)
		return
	}

	req := &aihub.CreateChatCompletionReq{
		Messages: []*aihub.Message{
			{
				Content: "请你评价以下AI Agent技术的现状，不超过200字",
				Role:    aihub.MessageRoleUser,
			},
		},
		Model: "gpt-3.5-turbo",
	}

	CreateChatCompletion(ctx, myLLM, req)
	// CreateChatCompletionStream(ctx, myProvider, req)
}

func CreateChatCompletion(ctx context.Context, myLLM aihub.ILLM, req *aihub.CreateChatCompletionReq) {
	rsp, err := myLLM.CreateChatCompletion(ctx, req)

	fmt.Println(err)
	fmt.Println("=======================")
	fmt.Println(rsp.Choices[0])
}

func CreateChatCompletionStream(ctx context.Context, myLLM aihub.ILLM, req *aihub.CreateChatCompletionReq) {
	stream := myLLM.CreateChatCompletionStream(ctx, req)

	for stream.Next() {
		data := stream.Current()
		fmt.Printf(data.Choices[0].Delta.Content)
		time.Sleep(10 * time.Millisecond)
		// fmt.Println(data)
	}

	fmt.Println("\n======[Done]=======")
	fmt.Println(stream.Err())
}
