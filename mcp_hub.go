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
	"github.com/mvptianyu/aihub/jsonschema"
	"log"
	"sync"
	"time"
)

type mcpHub struct {
	cliMaps    map[string]*client.SSEMCPClient // svraddr => cli
	toolMaps   map[string][]ToolFunction       // svraddr => []ToolFunction
	fnNameMaps map[string]*client.SSEMCPClient // funcname => cli

	lock sync.RWMutex
}

func (m *mcpHub) cronUpdateTools() {
	// 定时更新tools
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		m.lock.Lock()
		defer m.lock.Unlock()
		for addr, cli := range m.cliMaps {
			if toolFunctions, err1 := m.updateTools(cli, addr); err1 == nil {
				m.toolMaps[addr] = toolFunctions
				for _, toolFunction := range toolFunctions {
					m.fnNameMaps[toolFunction.Name] = cli
				}
			}
		}
	}
}

func (m *mcpHub) updateTools(cli *client.SSEMCPClient, addr string) (toolFunctions []ToolFunction, err error) {
	var toolRes *mcp.ListToolsResult
	toolRes, err = cli.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		log.Printf("RegisterMCPService::ListTools failed => addr:%s, err:%v\n", addr, err)
		return nil, err
	}

	toolFunctions = make([]ToolFunction, 0)

	// 转换tool
	for _, tool := range toolRes.Tools {
		toolFunction := ToolFunction{
			Parameters: &jsonschema.Definition{},
		}
		toolFunction.Name = tool.Name
		toolFunction.Description = tool.Description

		bs, _ := json.Marshal(tool.InputSchema)
		json.Unmarshal(bs, toolFunction.Parameters)
		toolFunctions = append(toolFunctions, toolFunction)
	}

	return toolFunctions, err
}

func (m *mcpHub) GetToolFunctions(addrs ...string) []ToolFunction {
	m.lock.RLock()
	defer m.lock.RUnlock()

	result := make([]ToolFunction, 0)
	for _, addr := range addrs {
		if m.toolMaps[addr] != nil {
			result = append(result, m.toolMaps[addr]...)
		}
	}
	return result
}

// 代理MCP请求
func (c *mcpHub) ProxyCall(ctx context.Context, name string, input string, output *Message) (rsp *mcp.CallToolResult, err error) {
	c.lock.RLock()
	cli := c.fnNameMaps[name]
	c.lock.RUnlock()
	if cli == nil {
		return nil, ErrMCPClientNotMatch
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
		if tmp, ok := h.cliMaps[addr]; ok {
			ret = append(ret, tmp)
		}
	}
	return ret
}

func (m *mcpHub) SetClient(addrs ...string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var err error
	var cli *client.SSEMCPClient
	ctx := context.Background()

	// 注册单例
	for _, addr := range addrs {
		if m.cliMaps[addr] != nil {
			continue
		}

		if cli, err = client.NewSSEMCPClient(addr); err != nil {
			log.Printf("SetClient::NewSSEMCPClient failed => addr:%s, err:%v\n", addr, err)
			return err
		}

		if err = cli.Start(ctx); err != nil {
			log.Printf("SetClient::Start failed => addr:%s, err:%v\n", addr, err)
			return err
		}

		// 初始化
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "aihub-mcp-client",
			Version: "1.0.0",
		}
		if _, err = cli.Initialize(ctx, initRequest); err != nil {
			log.Printf("SetClient::Initialize failed => addr:%s, err:%v\n", addr, err)
			return err
		}

		toolFunctions, err1 := m.updateTools(cli, addr)
		if err1 != nil {
			return err1
		}

		m.cliMaps[addr] = cli
		m.toolMaps[addr] = toolFunctions
		for _, toolFunction := range toolFunctions {
			m.fnNameMaps[toolFunction.Name] = cli
		}
	}

	return nil
}
