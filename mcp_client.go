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

type mcpClient struct {
	*client.SSEMCPClient
	toolFuncMaps map[string]ToolFunction
	addr         string
	timeout      time.Duration
	initAt       time.Time

	lock sync.RWMutex
}

const defaultMCPClientTimeout = 60 * time.Second

func newMCPClient(addr string) (*mcpClient, error) {
	now := time.Now()
	cli, err := client.NewSSEMCPClient(addr, client.WithSSEReadTimeout(defaultMCPClientTimeout))
	ctx := context.Background()
	if err != nil {
		log.Printf("newMCPClient::NewSSEMCPClient failed => addr:%s, err:%v\n", addr, err)
		return nil, err
	}

	if err = cli.Start(ctx); err != nil {
		log.Printf("newMCPClient::Start failed => addr:%s, err:%v\n", addr, err)
		return nil, err
	}

	// 初始化
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "aihub-mcp-client",
		Version: "1.0.0",
	}
	if _, err = cli.Initialize(ctx, initRequest); err != nil {
		log.Printf("newMCPClient::Initialize failed => addr:%s, err:%v\n", addr, err)
		return nil, err
	}

	ret := &mcpClient{
		SSEMCPClient: cli,
		toolFuncMaps: make(map[string]ToolFunction),
		addr:         addr,
		timeout:      defaultMCPClientTimeout,
		initAt:       now,
	}

	ret.updateTools()
	return ret, nil
}

func (m *mcpClient) updateTools() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	toolRes, err := m.SSEMCPClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Printf("mcpClient::updateTools failed => addr:%s, err:%v\n", m.addr, err)
		return false
	}

	// 转换tool
	for _, tool := range toolRes.Tools {
		toolFunction := ToolFunction{
			Parameters: &jsonschema.Definition{},
		}
		toolFunction.Name = tool.Name
		toolFunction.Description = tool.Description

		bs, _ := json.Marshal(tool.InputSchema)
		json.Unmarshal(bs, toolFunction.Parameters)

		// 额外加入SessionKey参数
		toolFunction.Parameters.Properties[ToolArgumentsRawSessionKey] = jsonschema.Definition{
			Type:        jsonschema.String,
			Description: "可选，声明工具运行结果写入到session数据的key名",
		}
		m.toolFuncMaps[toolFunction.Name] = toolFunction
	}
	return true
}

func (m *mcpClient) CheckValid() error {
	now := time.Now()
	if now.Sub(m.initAt) < m.timeout-5*time.Second {
		return nil
	}

	newCli, err := newMCPClient(m.addr)
	if err != nil {
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.SSEMCPClient = newCli.SSEMCPClient
	m.initAt = newCli.initAt
	m.toolFuncMaps = newCli.toolFuncMaps
	m.addr = newCli.addr
	m.timeout = newCli.timeout
	return nil
}

func (m *mcpClient) GetToolFunctions() []ToolFunction {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ret := make([]ToolFunction, 0)
	for _, tool := range m.toolFuncMaps {
		ret = append(ret, tool)
	}
	return ret
}
