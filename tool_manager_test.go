package aihub

import (
	"context"
	"testing"
)

// --------------------

type DelegateA struct {
}

type DelegateTestInput1 struct {
	ToolInputBase

	AA int `json:"aa"`
}

type DelegateTestInput2 struct {
	BB string `json:"bb"`
}

func (d *DelegateA) Method1(ctx context.Context, input *DelegateTestInput1, output *Message) (err error) {
	return nil
}

func (d *DelegateA) Method2(ctx context.Context, input *DelegateTestInput2, output *Message) (err error) {
	return nil
}

func TestToolManager_RegisterToolFunc(t *testing.T) {
	m := &ToolManager{
		toolMethods: make(map[string]*ToolMethod),
	}

	ctx := context.TODO()
	var err error
	delegate := &DelegateA{}
	if err = m.RegisterToolFunc(delegate); err != nil {
		t.Fatal(err)
	}

	call := &MessageToolCall{
		Id: "testID",
	}
	call.Function.Name = "Method1"
	call.Function.Arguments = "{\"aa\":1}"

	msg := &Message{}
	if err = m.InvokeToolFunc(ctx, call, msg); err != nil {
		t.Fatal(err)
	}
}
