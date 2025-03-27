package main

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/core"
)

func main() {
	ctx := context.Background()

	myProvider := aihub.NewProvider(&core.ProviderConfig{
		Name:    "openai",
		Model:   "gpt-3.5-turbo",
		BaseURL: "https://api.openai.com",
	})

	req := &core.CreateChatCompletionReq{
		Messages: []*core.Message{
			{
				Content: "1+1=?",
				Role:    core.MessageRoleUser,
			},
		},
		Model:  "gpt-3.5-turbo",
		Stream: false,
	}
	rsp, err := myProvider.CreateChatCompletion(ctx, req)

	fmt.Println(err)
	fmt.Println("=======================")
	fmt.Println(rsp.Choices[0])
}
