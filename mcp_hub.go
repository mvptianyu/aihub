package aihub

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"sync"
)

type mcpHub struct {
	clientMaps map[string]*mcpClient // svraddr => client
	fnMaps     map[string]*mcpClient // funcName => client

	lock sync.RWMutex
}

func (m *mcpHub) GetAllNameList() []string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ret := make([]string, 0)
	for name, _ := range m.clientMaps {
		ret = append(ret, name)
	}
	return ret
}

func (m *mcpHub) GetToolFunctions(addrs []string, names []string) []ToolFunction {
	m.lock.RLock()
	defer m.lock.RUnlock()

	result := make([]ToolFunction, 0)
	for _, addr := range addrs {
		// serverAddr过滤
		cli, ok := m.clientMaps[addr]
		if !ok || cli.CheckValid() != nil {
			continue
		}

		if names == nil || len(names) <= 0 {
			// 不过滤返回所有
			result = append(result, cli.GetToolFunctions()...)
			continue
		}

		// toolName过滤
		for _, name := range names {
			if tmp, ok2 := cli.toolFuncMaps[name]; ok2 {
				result = append(result, tmp)
			}
		}
	}
	return result
}

// ProxyCall 代理MCP请求
func (m *mcpHub) ProxyCall(ctx context.Context, name string, input string, output *Message) (err error) {
	m.lock.RLock()
	cli := m.fnMaps[name]
	m.lock.RUnlock()
	if cli == nil {
		err = ErrCallNameNotMatch
		return
	}
	if err = cli.CheckValid(); err != nil {
		return
	}

	args := make(map[string]interface{})
	if err = json.Unmarshal([]byte(input), &args); err != nil {
		return
	}

	request := mcp.CallToolRequest{}
	request.Params.Name = name
	request.Params.Arguments = args
	var rsp *mcp.CallToolResult
	rsp, err = cli.CallTool(ctx, request)
	if err != nil {
		return
	}

	if len(rsp.Content) < 1 {
		err = ErrMCPResponseEmpty
		return
	}

	if len(rsp.Content) == 1 {
		if textContent, ok := rsp.Content[0].(mcp.TextContent); ok {
			output.Content = textContent.Text
			return
		}
	}

	output.MultiContent = make([]*MessageContentPart, len(rsp.Content))
	for idx, content := range rsp.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			output.MultiContent[idx] = &MessageContentPart{
				Type: MessageContentTypeText,
				Text: textContent.Text,
			}
		} else if imageContent, ok := content.(mcp.ImageContent); ok {
			output.MultiContent[idx] = &MessageContentPart{
				Type: MessageContentTypeImage,
				ImageUrl: &MessageContentImage{
					URL: imageContent.Data,
				},
			}
		}
	}

	if len(output.MultiContent) < 1 {
		err = ErrMCPResponseEmpty
	}

	return
}

func (m *mcpHub) GetClient(addrs ...string) []*client.SSEMCPClient {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ret := make([]*client.SSEMCPClient, 0)
	for _, addr := range addrs {
		if cli, ok := m.clientMaps[addr]; ok {
			err := cli.CheckValid()
			if err != nil {
				ret = append(ret, nil)
			}
			ret = append(ret, cli.SSEMCPClient)
		}
	}
	return ret
}

func (m *mcpHub) SetClient(addrs ...string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// 注册单例
	for _, addr := range addrs {
		cli, err := newMCPClient(addr)
		if err != nil {
			continue
		}

		m.clientMaps[addr] = cli
		for _, toolFunction := range cli.GetToolFunctions() {
			m.fnMaps[toolFunction.Name] = cli
		}
	}

	return nil
}

func (m *mcpHub) DelClient(addrs ...string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, addr := range addrs {
		if cli, ok := m.clientMaps[addr]; ok {
			for _, toolFunction := range cli.GetToolFunctions() {
				delete(m.fnMaps, toolFunction.Name)
			}
		}
		delete(m.clientMaps, addr)
	}
	return nil
}

func (m *mcpHub) ConvertToOPENAPIConfig() string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	cfg := OPENAPIConfig{
		OpenAPI: "3.0.0",
		Info: OPENAPIInfo{
			Title:       "MCPHub's API Document",
			Description: "Generate by AIHub",
			Version:     "1.0.0",
		},
		Paths: make(map[string]OPENAPIPathItem),
		Tags:  make([]OPENAPITag, 0),
	}

	addrs := m.GetAllNameList()
	for _, addr := range addrs {
		cfg.Tags = append(cfg.Tags, OPENAPITag{
			Name: addr,
		})
	}

	for server, item := range m.clientMaps {
		err := item.CheckValid()
		if err != nil {
			continue
		}

		cfg.AddToolFunction(item.GetToolFunctions(), server)
	}

	bs, _ := json.Marshal(cfg)
	return string(bs)
}
