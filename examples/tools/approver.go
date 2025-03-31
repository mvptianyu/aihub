/*
@Project: aihub
@Module: tools
@File : approve.go
*/
package tools

import "context"

type Approver struct {
}

// SubmitApplication 提交授权申请
func (m *Approver) BeforeProcessing(ctx context.Context, question string, timeout int64, session map[string]interface{}) error {
	return nil
}

// SubmitApplication 提交授权申请
func (m *Approver) OnProcessing(ctx context.Context, question string, timeout int64, session map[string]interface{}) error {
	return nil
}

// SubmitApplication 提交授权申请
func (m *Approver) AfterProcessing(ctx context.Context, question string, timeout int64, session map[string]interface{}) error {
	return nil
}
