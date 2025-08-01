package aihub

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"os"
	"path/filepath"
	"sync"
)

type llmHub struct {
	llms map[string]ILLM // Name => ILLM

	mcpSrv IMCPServer
	lock   sync.RWMutex
}

func (h *llmHub) GetAllNameList() []string {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]string, 0)
	for name, _ := range h.llms {
		ret = append(ret, name)
	}
	return ret
}

func (h *llmHub) GetLLMList(names ...string) []ILLM {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]ILLM, 0)
	for _, name := range names {
		if tmp, ok := h.llms[name]; ok {
			ret = append(ret, tmp)
		}
	}
	return ret
}

func (h *llmHub) GetLLM(name string) ILLM {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if tmp, ok := h.llms[name]; ok {
		return tmp
	}
	return nil
}

func (h *llmHub) DelLLM(name string) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.llms, name)
	h.delMCPServerTool(name)
	return nil
}

func (h *llmHub) SetLLM(cfg *LLMConfig) (ILLM, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, ok := h.llms[cfg.Name]; ok {
		delete(h.llms, cfg.Name) // 删除旧的
	}

	ins, err := newLLM(cfg)
	if err != nil {
		return nil, err
	}
	h.llms[cfg.Name] = ins
	h.addMCPServerTool(ins) // 加入MCPServer
	return ins, err
}

func (h *llmHub) SetLLMByYamlData(yamlData []byte) (ILLM, error) {
	cfg, err := YamlDataToLLMConfig(yamlData)
	if err != nil {
		return nil, err
	}
	return h.SetLLM(cfg)
}

func (h *llmHub) SetLLMByYamlFile(yamlFile string) (ILLM, error) {
	// 读取 YAML 文件内容
	yamlData, err := os.ReadFile(filepath.Clean(yamlFile))
	if err != nil {
		fmt.Printf("Error reading YAML file: %s => %v\n", yamlFile, err)
		return nil, err
	}
	return h.SetLLMByYamlData(yamlData)
}

func (h *llmHub) addMCPServerTool(item ILLM) {
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
				return rsp, fmt.Errorf("empty param: INPUT_")
			}

			innerReq := &CreateChatCompletionReq{
				Messages: []*Message{
					{
						Role:    MessageRoleUser,
						Content: innerInput,
					},
				},
			}
			innerRsp, err := item.CreateChatCompletion(ctx, innerReq)
			if err == nil && innerRsp.Choices != nil && len(innerRsp.Choices) > 0 {
				choice := innerRsp.Choices[0]
				rsp.Content = append(rsp.Content, mcp.NewTextContent(choice.Message.Content))
				rsp.IsError = false
			}

			return rsp, err
		},
	}
	h.mcpSrv.AddTools(tool)
}

func (h *llmHub) delMCPServerTool(name string) {
	if h.mcpSrv == nil {
		return
	}

	h.mcpSrv.DelTools(name)
}

func (a *llmHub) GetMCPServer() IMCPServer {
	return a.mcpSrv
}
