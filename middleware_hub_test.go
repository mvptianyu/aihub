/*
@Project: aihub
@Module: aihub
@File : middleware_hub_test.go
*/
package aihub

import (
	"context"
	"fmt"
	"testing"
)

type DemoMiddleware struct{}

func (m *DemoMiddleware) BeforeProcessing(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) error {
	fmt.Printf("===> BeforeProcessing toolCalls: %v, sessionData: %v\n", toolCalls, opts.RuntimeCfg.SessionData)
	return nil
}

// SubmitApplication 提交授权申请
func (m *DemoMiddleware) AfterProcessing(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) error {
	fmt.Printf("===> AfterProcessing toolCalls: %v, sessionData: %v\n", toolCalls, opts.RuntimeCfg.SessionData)
	return nil
}

func Test_middlewareHub_SetMiddleware(t *testing.T) {
	err := GetMiddlewareHub().SetMiddleware(&DemoMiddleware{})
	fmt.Println(err)
}
