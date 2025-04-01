/*
@Project: aihub
@Module: core
@File : mcp.go
*/
package core

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mvptianyu/aihub/jsonschema"
	"log"
	"sync"
	"time"
)

type mcpManager struct {
	clients   map[string]*client.SSEMCPClient // svraddr => cli
	tools     map[string][]ToolFunction       // svraddr => []ToolFunction
	fnCliMaps map[string]*client.SSEMCPClient // funcname => cli

	lock sync.RWMutex
}

var defaultMCPManager *mcpManager
var defaultMCPManagerOnce sync.Once

func GetDefaultMCPManager() *mcpManager {
	defaultMCPManagerOnce.Do(func() {
		defaultMCPManager = &mcpManager{
			clients:   make(map[string]*client.SSEMCPClient),
			tools:     make(map[string][]ToolFunction),
			fnCliMaps: make(map[string]*client.SSEMCPClient),
		}
		go defaultMCPManager.cronUpdateTools()
	})
	return defaultMCPManager
}

func (m *mcpManager) RegisterMCPService(mcpServerAddrs ...string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var err error
	var cli *client.SSEMCPClient
	ctx := context.Background()

	// 注册单例
	for _, mcpServerAddr := range mcpServerAddrs {
		if m.clients[mcpServerAddr] != nil {
			continue
		}

		if cli, err = client.NewSSEMCPClient(mcpServerAddr); err != nil {
			log.Printf("RegisterMCPService::NewSSEMCPClient failed => server:%s, err:%v\n", mcpServerAddr, err)
			return err
		}

		if err = cli.Start(ctx); err != nil {
			log.Printf("RegisterMCPService::Start failed => server:%s, err:%v\n", mcpServerAddr, err)
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
			log.Printf("RegisterMCPService::Initialize failed => server:%s, err:%v\n", mcpServerAddr, err)
			return err
		}

		toolFunctions, err1 := m.updateTools(cli, mcpServerAddr)
		if err1 != nil {
			return err1
		}

		m.clients[mcpServerAddr] = cli
		m.tools[mcpServerAddr] = toolFunctions
		for _, toolFunction := range toolFunctions {
			m.fnCliMaps[toolFunction.Name] = cli
		}
	}

	return nil
}

func (m *mcpManager) cronUpdateTools() {
	// 定时更新tools
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		m.lock.Lock()
		defer m.lock.Unlock()
		for mcpServerAddr, cli := range m.clients {
			if toolFunctions, err1 := m.updateTools(cli, mcpServerAddr); err1 == nil {
				m.tools[mcpServerAddr] = toolFunctions
				for _, toolFunction := range toolFunctions {
					m.fnCliMaps[toolFunction.Name] = cli
				}
			}
		}
	}
}

func (m *mcpManager) updateTools(cli *client.SSEMCPClient, mcpServerAddr string) (toolFunctions []ToolFunction, err error) {
	var toolRes *mcp.ListToolsResult
	toolRes, err = cli.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		log.Printf("RegisterMCPService::ListTools failed => server:%s, err:%v\n", mcpServerAddr, err)
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

func (m *mcpManager) GetToolFunctions(mcpServerAddrs ...string) []ToolFunction {
	m.lock.RLock()
	defer m.lock.RUnlock()

	result := make([]ToolFunction, 0)
	for _, mcpServerAddr := range mcpServerAddrs {
		if m.tools[mcpServerAddr] != nil {
			result = append(result, m.tools[mcpServerAddr]...)
		}
	}
	return result
}

// 代理请求
func (c *mcpManager) ProxyMCPCall(ctx context.Context, input *ToolInputBase, output *Message) (err error) {
	params := make(map[string]interface{})
	if err = json.Unmarshal([]byte(input.GetRawInput()), &params); err != nil {
		return err
	}

	funcName := input.GetRawFuncName()
	c.lock.RLock()
	cli := c.fnCliMaps[funcName]
	c.lock.RUnlock()

	if cli == nil {
		return errors.New("not found matched mcp client instance")
	}

	request := mcp.CallToolRequest{}
	request.Params.Name = funcName
	request.Params.Arguments = params
	result, err := cli.CallTool(ctx, request)
	if err != nil {
		return err
	}
	if len(result.Content) < 1 {
		return errors.New("mcp response empty error")
	}

	if len(result.Content) == 1 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			output.Content = textContent.Text
			return nil
		}
	}

	output.MultiContent = make([]*MessageContentPart, len(result.Content))
	for idx, content := range result.Content {
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
		return errors.New("mcp response no valid text/image content error")
	}
	return nil
}
