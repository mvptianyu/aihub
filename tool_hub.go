package aihub

import (
	"context"
	"encoding/json"
	"github.com/mvptianyu/aihub/jsonschema"
	"github.com/tidwall/gjson"
	"log"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

// ToolMethod 工具方法入口签名，Input派生自ToolInputBase
type ToolMethod func(ctx context.Context, input IToolInput, output *Message) (err error)

type ToolEntry struct {
	Description string
	Function    interface{} // 方法入口

	method       reflect.Value
	input        reflect.Type
	toolFunction ToolFunction
}

type toolHub struct {
	toolEntrys map[string]ToolEntry // toolFunc Name => toolFunc

	lock sync.RWMutex
}

func (h *toolHub) GetAllNameList() []string {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]string, 0)
	for name, _ := range h.toolEntrys {
		ret = append(ret, name)
	}
	return ret
}

func (h *toolHub) GetTool(names ...string) []ToolEntry {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]ToolEntry, 0)
	for _, name := range names {
		if tmp, ok := h.toolEntrys[name]; ok {
			ret = append(ret, tmp)
		}
	}
	return ret
}

func (h *toolHub) GetToolFunctions(names ...string) []ToolFunction {
	h.lock.RLock()
	defer h.lock.RUnlock()

	result := make([]ToolFunction, 0)
	for _, name := range names {
		if tmp, ok := h.toolEntrys[name]; ok {
			result = append(result, tmp.toolFunction)
		}
	}
	return result
}

func (h *toolHub) SetTool(objs ...ToolEntry) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	var err error
	for _, obj := range objs {
		if obj.Function == nil {
			continue
		}
		obj.method = reflect.ValueOf(obj.Function)
		// 获取函数名
		methodName := runtime.FuncForPC(obj.method.Pointer()).Name()
		splits := strings.Split(methodName, ".")
		fixName := splits[len(splits)-1] // 去除函数名中的包路径

		// 获取函数类型
		methodType := reflect.TypeOf(obj.Function)
		// 重复注册
		if _, ok := h.toolEntrys[fixName]; ok {
			return ErrToolRegisterRepeat
		}

		// 检查方法的出入参数量
		if methodType.NumIn() != 4 && methodType.NumOut() != 1 {
			continue
		}

		// 检查第一个参数是否为 Context.Context 类型
		if methodType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
			continue
		}

		// 检查第二个参数是否实现了 IToolInput 接口
		if methodType.In(1).Implements(reflect.TypeOf((*IToolInput)(nil)).Elem()) == false {
			continue
		}

		// 检查第三个返回值是否为 *Message 类型
		if methodType.In(2) != reflect.TypeOf((*Message)(nil)) {
			continue
		}

		// 检查第1个返回值是否为 error 类型
		if methodType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		// 获取第二个参数的类型（索引为 1，因为第一个是 ctx）
		obj.input = methodType.In(1).Elem()
		obj.toolFunction = ToolFunction{}
		obj.toolFunction.Name = fixName
		if obj.Description == "" {
			obj.Description = fixName
		}
		obj.toolFunction.Description = obj.Description

		if obj.toolFunction.Parameters, err = jsonschema.GenerateSchemaForType(obj.input); err != nil {
			log.Printf("jsonschema.GenerateSchemaForType failed => name:%s, err:%v\n", methodName, err)
			continue
		}

		if obj.toolFunction.Parameters.Properties == nil {
			obj.toolFunction.Parameters.Properties = make(map[string]jsonschema.Definition)
		}
		if len(obj.toolFunction.Parameters.Properties) == 0 {
			obj.toolFunction.Parameters.Properties[ToolArgumentsRawInputKey] = jsonschema.Definition{
				Type:        jsonschema.String,
				Description: ToolArgumentsRawInputKey,
			}
		}

		h.toolEntrys[fixName] = obj
	}
	return nil
}

func (h *toolHub) DelTool(names ...string) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	for _, name := range names {
		delete(h.toolEntrys, name)
	}
	return nil
}

// ProxyCall 代理ToolCall请求
func (h *toolHub) ProxyCall(ctx context.Context, name string, input string, output *Message) (err error) {
	tmpToolEntrys := h.GetTool(name)
	if tmpToolEntrys == nil || len(tmpToolEntrys) <= 0 {
		return ErrCallNameNotMatch
	}
	toolEntry := tmpToolEntrys[0]

	// 获取结构体实例的反射值
	inputValue := reflect.New(toolEntry.input)
	if err = json.Unmarshal([]byte(input), inputValue.Interface()); err != nil {
		return err
	}
	if rawArgs := gjson.Get(input, ToolArgumentsRawInputKey).String(); rawArgs != "" {
		inputValue.Interface().(IToolInput).SetRawInput(rawArgs)
	}

	// 反射调用函数
	results := toolEntry.method.Call([]reflect.Value{
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

func (h *toolHub) ConvertToOPENAPIConfig() string {
	h.lock.RLock()
	defer h.lock.RUnlock()

	cfg := OPENAPIConfig{
		OpenAPI: "3.0.0",
		Info: OPENAPIInfo{
			Title:       "ToolHub's API Document",
			Description: "Generate by AIHub",
			Version:     "1.0.0",
		},
		Paths: make(map[string]OPENAPIPathItem),
	}

	toolFunctions := make([]ToolFunction, 0)
	for _, item := range h.toolEntrys {
		toolFunctions = append(toolFunctions, item.toolFunction)
	}
	cfg.AddToolFunction(toolFunctions, "")
	bs, _ := json.Marshal(cfg)
	return string(bs)
}
