package core

import (
	"context"
	"encoding/json"
	"github.com/mvptianyu/aihub/jsonschema"
	"reflect"
	"sync"
)

// ToolFunc 工具方法签名，Input派生自ToolInputBase
type ToolFunc func(ctx context.Context, input IToolInput, output *Message) (err error)

type ToolItem struct {
	delegate reflect.Value
	method   reflect.Method
	input    reflect.Type
	schema   *jsonschema.Definition
}

type ToolManager struct {
	tools map[string]*ToolItem

	lock sync.RWMutex
}

func (m *ToolManager) RegisterToolFunc(delegate interface{}) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	delegateType := reflect.TypeOf(delegate)
	delegateVal := reflect.ValueOf(delegate)
	var err error

	for i := 0; i < delegateType.NumMethod(); i++ {
		method := delegateType.Method(i)
		methodType := method.Type

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

		// 加入
		if _, ok := m.tools[methodType.Name()]; ok {
			return ErrToolRegisterRepeat
		}

		item := &ToolItem{
			delegate: delegateVal,
			method:   method,
			input:    methodType.In(2).Elem(), // struct类型
		}

		if item.schema, err = jsonschema.GenerateSchemaForType(item.input); err != nil {
			return err
		}

		m.tools[method.Name] = item
	}
	return nil
}

// 反射调用指定名称的方法
func (m *ToolManager) InvokeToolFunc(name string, ctx context.Context, input string, output *Message) error {
	m.lock.RLock()
	item, ok := m.tools[name]
	m.lock.RUnlock()
	if !ok {
		return ErrToolRegisterEmpty
	}

	var err error
	// 获取结构体实例的反射值
	inputValue := reflect.New(item.input)
	if err = json.Unmarshal([]byte(input), inputValue.Interface()); err != nil {
		return err
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
