/*
@Project: aihub
@Module: tools
@File : approve.go
*/
package tools

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/mvptianyu/aihub/core"
	"os"
	"strings"
	"sync"
)

type Approver struct {
	approveMap map[string]chan bool // requestId => true:同意，false:失败

	lock sync.RWMutex
}

const seatalkGroup = "LMWNqAYCQVGLGi2fGYfvHw"

// SubmitApplication 提交授权申请
func (m *Approver) BeforeProcessing(ctx context.Context, toolCalls []*core.MessageToolCall, opts *core.RunOptions) error {
	fmt.Printf("===> BeforeProcessing toolCalls: %v, sessionData: %v\n", toolCalls, opts.RuntimeCfg.SessionData)
	requestID := ""
	for _, call := range toolCalls {
		requestID += "|" + call.Id
	}
	requestID = strings.TrimLeft(requestID, "|")

	m.lock.Lock()
	if m.approveMap == nil {
		m.approveMap = make(map[string]chan bool)
	}
	m.approveMap[requestID] = make(chan bool)
	m.lock.Unlock()

	go func() {
		// 发审批请求
		core.SendSeatalkText(seatalkGroup, core.SeaTalkText{
			Content: "即将调用工具ID: " + requestID + "，是否同意？",
		})

		m.OnProcessing(ctx, toolCalls, opts)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case result := <-m.approveMap[requestID]:
		if !result {
			return errors.New("ToolCall reject by user")
		}
	}

	return nil
}

// SubmitApplication 提交授权申请
func (m *Approver) OnProcessing(ctx context.Context, toolCalls []*core.MessageToolCall, opts *core.RunOptions) error {
	fmt.Printf("===> OnProcessing toolCalls: %v, sessionData: %v\n", toolCalls, opts.RuntimeCfg.SessionData)
	requestID := ""
	for _, call := range toolCalls {
		requestID += "|" + call.Id
	}
	requestID = strings.TrimLeft(requestID, "|")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	userInput := scanner.Text()

	m.lock.Lock()
	defer m.lock.Unlock()

	ret := false
	if userInput == "OK" {
		ret = true
	}
	m.approveMap[requestID] <- ret

	return nil
}

// SubmitApplication 提交授权申请
func (m *Approver) AfterProcessing(ctx context.Context, toolCalls []*core.MessageToolCall, opts *core.RunOptions) error {
	fmt.Printf("===> AfterProcessing toolCalls: %v, sessionData: %v\n", toolCalls, opts.RuntimeCfg.SessionData)
	return nil
}
