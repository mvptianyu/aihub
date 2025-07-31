package aihub

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type agentHub struct {
	agents map[string]IAgent // agent Name => ILLM

	mcpSrv IMCPServer
	lock   sync.RWMutex
}

func (h *agentHub) GetAllNameList() []string {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]string, 0)
	for name, _ := range h.agents {
		ret = append(ret, name)
	}
	return ret
}

func (h *agentHub) GetAgentList(names ...string) []IAgent {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]IAgent, 0)
	for _, name := range names {
		if tmp, ok := h.agents[name]; ok {
			ret = append(ret, tmp)
		}
	}
	return ret
}

func (h *agentHub) GetAgent(name string) IAgent {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if tmp, ok := h.agents[name]; ok {
		return tmp
	}
	return nil
}

func (h *agentHub) DelAgent(name string) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.agents, name)
	h.delMCPServerTool(name)
	return nil
}

func (h *agentHub) SetAgent(cfg *AgentConfig) (IAgent, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	ag, err := newAgent(cfg)
	if err != nil {
		return nil, err
	}
	h.agents[cfg.Name] = ag
	h.addMCPServerTool(ag) // 加入MCPServer
	return ag, err
}

func (h *agentHub) SetAgentByYamlData(yamlData []byte) (IAgent, error) {
	cfg, err := YamlDataToAgentConfig(yamlData)
	if err != nil {
		return nil, err
	}
	return h.SetAgent(cfg)
}

func (h *agentHub) SetAgentByYamlFile(yamlFile string) (IAgent, error) {
	// 读取 YAML 文件内容
	yamlData, err := os.ReadFile(filepath.Clean(yamlFile))
	if err != nil {
		log.Printf("Error reading YAML file: %s => %v\n", yamlFile, err)
		return nil, err
	}
	return h.SetAgentByYamlData(yamlData)
}

func (h *agentHub) addMCPServerTool(item IAgent) {
	if h.mcpSrv == nil {
		return
	}

	briefInfo := item.GetBriefInfo()
	tool := server.ServerTool{
		Tool: mcp.Tool{
			Name:        briefInfo.Name,
			Description: briefInfo.Description,
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					ToolArgumentsRawInputKey: map[string]interface{}{
						"type":        "string",
						"description": "原封不动传递的用户提问语句",
					},
				},
				Required: []string{
					ToolArgumentsRawInputKey,
				},
			},
		},
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			rsp := &mcp.CallToolResult{
				Content: make([]mcp.Content, 0),
				IsError: true,
			}
			innerInput := ""
			if tmpStr, ok := request.Params.Arguments[ToolArgumentsRawInputKey]; ok {
				innerInput, _ = tmpStr.(string)
			}
			if innerInput == "" {
				return rsp, fmt.Errorf("empty param: _INPUT_")
			}

			innerRsp := item.Run(ctx, innerInput)
			if innerRsp.Err == nil {
				rsp.Content = append(rsp.Content, mcp.NewTextContent(innerRsp.Content))
				rsp.IsError = false
			}

			return rsp, innerRsp.Err
		},
	}
	h.mcpSrv.AddTools(tool)
}

func (h *agentHub) delMCPServerTool(name string) {
	if h.mcpSrv == nil {
		return
	}

	h.mcpSrv.DelTools(name)
}

func (a *agentHub) GetMCPServer() IMCPServer {
	return a.mcpSrv
}
