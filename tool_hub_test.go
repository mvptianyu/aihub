/*
@Project: aihub
@Module: aihub
@File : toolentry_hub_test.go
*/
package aihub

import (
	"context"
	"testing"
)

type Method1Input struct {
	ToolInputBase

	AA int `json:"aa"`
}

type Method2Input struct {
	BB string `json:"bb"`
}

func Method1(ctx context.Context, input *Method1Input, output *Message) (err error) {
	return nil
}

func Method2(ctx context.Context, input *Method2Input, output *Message) (err error) {
	return nil
}

func Test_toolEntryHub_SetToolEntry(t *testing.T) {
	ctx := context.Background()
	var err error

	GetToolHub().SetTool(
		ToolEntry{
			Function:    Method1,
			Description: "Method1 desc",
		},
		ToolEntry{
			Function:    Method2,
			Description: "Method2 desc",
		},
	)

	msg := &Message{}
	if err = GetToolHub().ProxyCall(ctx, "Method1", "{\"aa\":333}", msg); err != nil {
		t.Fatal(err)
	}
}
