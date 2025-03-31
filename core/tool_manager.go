package core

import (
	"context"
	"encoding/json"
	"github.com/mvptianyu/aihub/jsonschema"
	"github.com/tidwall/gjson"
	"reflect"
	"sync"
)

// ToolFunc 工具方法签名，Input派生自ToolInputBase
type ToolFunc func(ctx context.Context, input IToolInput, output *Message) (err error)

type ToolMethod struct {
	name     string
	delegate reflect.Value
	method   reflect.Method
	input    reflect.Type
}

type ToolManager struct {
	toolMethods   map[string]*ToolMethod
	toolFunctions []ToolFunction
	middlewares   []IMiddleware

	lock sync.RWMutex
}

func NewToolManager() *ToolManager {
	return &ToolManager{
		toolMethods:   make(map[string]*ToolMethod),
		toolFunctions: make([]ToolFunction, 0),
	}
}

func (m *ToolManager) RegisterToolFunc(delegate interface{}, enableTools []ToolFunction) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	bCheckEnableTool := false // 是否需要校验启用列表
	enableToolMap := make(map[string]ToolFunction)
	if enableTools != nil && len(enableTools) > 0 {
		for _, enableTool := range enableTools {
			enableToolMap[enableTool.Name] = enableTool
			bCheckEnableTool = true
		}
	}

	delegateType := reflect.TypeOf(delegate)
	delegateVal := reflect.ValueOf(delegate)
	var err error

	for i := 0; i < delegateType.NumMethod(); i++ {
		method := delegateType.Method(i)
		methodType := method.Type

		// 重复注册
		if _, ok := m.toolMethods[method.Name]; ok {
			return ErrToolRegisterRepeat
		}

		// 检查方法的出入参数量
		if methodType.NumIn() != 4 && methodType.NumOut() != 1 {
			continue
		}

		// 检查第一个参数是否为 context.Context 类型
		if methodType.In(1) != reflect.TypeOf((*context.Context)(nil)).Elem() {
			continue
		}

		// 检查第二个参数是否实现了 IToolInput 接口
		if methodType.In(2).Implements(reflect.TypeOf((*IToolInput)(nil)).Elem()) == false {
			continue
		}

		// 检查第三个返回值是否为 *Message 类型
		if methodType.In(3) != reflect.TypeOf((*Message)(nil)) {
			continue
		}

		// 检查第1个返回值是否为 error 类型
		if methodType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		methodItem := &ToolMethod{
			name:     method.Name,
			delegate: delegateVal,
			method:   method,
			input:    methodType.In(2).Elem(), // struct类型
		}

		CfgItem := ToolFunction{}
		if bCheckEnableTool {
			ok := false
			if CfgItem, ok = enableToolMap[method.Name]; !ok {
				continue
			}
		}

		if CfgItem.Name == "" {
			CfgItem.Name = method.Name
		}
		if CfgItem.Description == "" {
			CfgItem.Description = CfgItem.Name
		}

		if CfgItem.Parameters, err = jsonschema.GenerateSchemaForType(methodItem.input); err != nil {
			return err
		}

		if CfgItem.Parameters.Properties == nil || len(CfgItem.Parameters.Properties) == 0 {
			CfgItem.Parameters.Properties[ToolArgumentsRawInputKey] = jsonschema.Definition{
				Type:        jsonschema.String,
				Description: ToolArgumentsRawInputKey,
			}
		}

		// 加入
		m.toolFunctions = append(m.toolFunctions, CfgItem)
		m.toolMethods[method.Name] = methodItem
	}
	return nil
}

func (m *ToolManager) GetToolDefinition() []ToolFunction {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.toolFunctions
}

func (m *ToolManager) GetToolCfg() []*Tool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ret := make([]*Tool, 0)
	for _, item := range m.toolFunctions {
		ret = append(ret, &Tool{
			Type:     ToolTypeFunction,
			Function: item,
		})
	}
	return ret
}

// InvokeToolFunc 反射调用指定名称的方法
func (m *ToolManager) InvokeToolFunc(ctx context.Context, toolCall *MessageToolCall, output *Message, opts *RunOptions) error {
	m.lock.RLock()
	item, ok := m.toolMethods[toolCall.Function.Name]
	m.lock.RUnlock()
	if !ok {
		return ErrToolRegisterEmpty
	}

	var err error
	// 获取结构体实例的反射值
	inputValue := reflect.New(item.input)
	if err = json.Unmarshal([]byte(toolCall.Function.Arguments), inputValue.Interface()); err != nil {
		return err
	}

	if rawArgs := gjson.Get(toolCall.Function.Arguments, ToolArgumentsRawInputKey).String(); rawArgs != "" {
		inputValue.Interface().(IToolInput).SetRawInput(rawArgs)
	}
	inputValue.Interface().(IToolInput).SetRawCallID(toolCall.Id)
	if opts != nil {
		inputValue.Interface().(IToolInput).SetRawSession(opts.session)
	}

	// 调用方法
	results := item.method.Func.Call([]reflect.Value{
		item.delegate,
		reflect.ValueOf(ctx),
		inputValue,
		reflect.ValueOf(output),
	})

	// 获取返回值
	errValue := results[0]
	if !errValue.IsNil() {
		err = errValue.Interface().(error)
	}

	return err
}

// ProcessToolCalls 处理本步骤tookcalls
func (m *ToolManager) ProcessToolCalls(ctx context.Context, toolCalls []*MessageToolCall, opts *RunOptions) (toolMsgs []*Message, err error) {

	// todo: 是否需要先触发拦截器（授权）
	if m.middlewares != nil {
		for _, middleware := range m.middlewares {
			middleware.BeforeProcessing()

		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(toolCalls))
	toolMsgs = make([]*Message, len(toolCalls))
	for i := 0; i < len(toolCalls); i++ {
		toolCall := toolCalls[i]
		toolMsgs[i] = &Message{
			Role:         MessageRoleTool,
			ToolCallID:   toolCall.Id,
			MultiContent: make([]*MessageContentPart, 0),
		}

		go func(i int, toolCall *MessageToolCall) {
			defer wg.Done()

			err1 := m.InvokeToolFunc(ctx, toolCall, toolMsgs[i], opts)
			if err1 != nil {
				err = err1
				return
			}
		}(i, toolCall)
	}
	wg.Wait()
	return
}
