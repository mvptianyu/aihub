/*
@Project: aihub
@Module: aihub
@File : provider_hub.go
*/
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

func (m *mcpHub) GetToolFunctions(addrs ...string) []ToolFunction {
	m.lock.RLock()
	defer m.lock.RUnlock()

	result := make([]ToolFunction, 0)
	for _, addr := range addrs {
		if client, ok := m.clientMaps[addr]; ok {
			err := client.CheckValid()
			if err != nil {
				continue
			}
			result = append(result, client.toolFunctions...)
		}
	}
	return result
}

// 代理MCP请求
func (c *mcpHub) ProxyCall(ctx context.Context, name string, input string, output *Message) (rsp *mcp.CallToolResult, err error) {
	c.lock.RLock()
	cli := c.fnMaps[name]
	c.lock.RUnlock()
	if cli == nil {
		return nil, ErrMCPClientNotMatch
	}
	if err = cli.CheckValid(); err != nil {
		return nil, err
	}

	args := make(map[string]interface{})
	if err = json.Unmarshal([]byte(input), &args); err != nil {
		return nil, err
	}

	request := mcp.CallToolRequest{}
	request.Params.Name = name
	request.Params.Arguments = args
	rsp, err = cli.CallTool(ctx, request)
	if err != nil {
		return nil, err
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

func (h *mcpHub) GetClient(addrs ...string) []*client.SSEMCPClient {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]*client.SSEMCPClient, 0)
	for _, addr := range addrs {
		if client, ok := h.clientMaps[addr]; ok {
			err := client.CheckValid()
			if err != nil {
				ret = append(ret, nil)
			}
			ret = append(ret, client.SSEMCPClient)
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
		for _, toolFunction := range cli.toolFunctions {
			m.fnMaps[toolFunction.Name] = cli
		}
	}

	return nil
}

func (h *mcpHub) DelClient(addrs ...string) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	for _, addr := range addrs {
		if cli, ok := h.clientMaps[addr]; ok {
			for _, toolFunction := range cli.toolFunctions {
				delete(h.fnMaps, toolFunction.Name)
			}
		}
		delete(h.clientMaps, addr)
	}
	return nil
}
