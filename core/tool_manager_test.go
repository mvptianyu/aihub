package core

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
		tools: make(map[string]*ToolItem),
	}

	ctx := context.TODO()
	var err error
	delegate := &DelegateA{}
	if err = m.RegisterToolFunc(delegate); err != nil {
		t.Fatal(err)
	}

	msg := &Message{}
	if err = m.InvokeToolFunc("Method1", ctx, "{\"aa\":1}", msg); err != nil {
		t.Fatal(err)
	}
}
