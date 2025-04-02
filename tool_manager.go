package aihub

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

type ToolMode int

const (
	ToolModeGeneral ToolMode = 0
	ToolModeMCP     ToolMode = 1
)

type ToolMethod struct {
	name     string
	delegate reflect.Value
	method   reflect.Method
	input    reflect.Type
	mode     ToolMode
}

type ToolManager struct {
	toolMethods   map[string]*ToolMethod
	toolFunctions []ToolFunction
	cfg           *AgentRuntimeCfg

	lock sync.RWMutex
}

func NewToolManager(cfg *AgentRuntimeCfg) IToolManager {
	return &ToolManager{
		toolMethods:   make(map[string]*ToolMethod),
		toolFunctions: make([]ToolFunction, 0),
		cfg:           cfg,
	}
}

// 注册MCP方法
func (m *ToolManager) RegisterMCPFunc() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.cfg.Mcps == nil || len(m.cfg.Mcps) < 1 {
		return nil
	}

	// 注册
	if err := GetDefaultMCPManager().RegisterMCPService(m.cfg.Mcps...); err != nil {
		return err
	}

	toolFunctions := GetDefaultMCPManager().GetToolFunctions(m.cfg.Mcps...)
	m.toolFunctions = append(m.toolFunctions, toolFunctions...)

	delegate := GetDefaultMCPManager()
	delegateType := reflect.TypeOf(delegate)
	delegateValue := reflect.ValueOf(delegate)
	inputType := reflect.TypeOf(ToolInputBase{})
	for _, toolFunction := range toolFunctions {
		toolMethod := &ToolMethod{
			name:     toolFunction.Name,
			delegate: delegateValue,
			input:    inputType,
			mode:     ToolModeMCP,
		}
		toolMethod.method, _ = delegateType.MethodByName("ProxyMCPCall")
		m.toolMethods[toolFunction.Name] = toolMethod
	}

	return nil
}

func (m *ToolManager) RegisterToolFunc(delegate interface{}) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	bCheckEnableTool := false // 是否需要校验启用列表
	enableToolMap := make(map[string]ToolSummary)
	enableTools := m.cfg.Tools
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
			mode:     ToolModeGeneral,
		}

		CfgItem := ToolFunction{}
		if bCheckEnableTool {
			ok := false
			summary, ok := enableToolMap[method.Name]
			if !ok {
				continue
			}
			CfgItem.ToolSummary = summary
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
func (m *ToolManager) InvokeToolFunc(ctx context.Context, toolCall *MessageToolCall, output *Message) error {
	m.lock.RLock()
	item, ok := m.toolMethods[toolCall.Function.Name]
	m.lock.RUnlock()
	if !ok {
		return ErrToolRegisterEmpty
	}

	var err error
	// 获取结构体实例的反射值
	inputValue := reflect.New(item.input)
	switch item.mode {
	case ToolModeGeneral:
		if err = json.Unmarshal([]byte(toolCall.Function.Arguments), inputValue.Interface()); err != nil {
			return err
		}
		if rawArgs := gjson.Get(toolCall.Function.Arguments, ToolArgumentsRawInputKey).String(); rawArgs != "" {
			inputValue.Interface().(IToolInput).SetRawInput(rawArgs)
		}
	case ToolModeMCP:
		inputValue.Interface().(IToolInput).SetRawInput(toolCall.Function.Arguments)
	}
	inputValue.Interface().(IToolInput).SetRawCallID(toolCall.Id)
	inputValue.Interface().(IToolInput).SetRawFuncName(toolCall.Function.Name)

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
