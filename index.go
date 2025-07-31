package aihub

import (
	"sync"
)

// ================AgentHub================
var defaultAgentHub *agentHub
var defaultAgentHubOnce sync.Once

func GetAgentHub() IAgentHub {
	defaultAgentHubOnce.Do(func() {
		defaultAgentHub = &agentHub{
			agents: make(map[string]IAgent),
			mcpSrv: newMCPServer("agent"),
		}
	})
	return defaultAgentHub
}

// ================LLMHub================
var defaultLLMHub *llmHub
var defaultLLMHubOnce sync.Once

func GetLLMHub() ILLMHub {
	defaultLLMHubOnce.Do(func() {
		defaultLLMHub = &llmHub{
			llms:   make(map[string]ILLM),
			mcpSrv: newMCPServer("llm"),
		}
	})
	return defaultLLMHub
}

// ================MCPHub================
var defaultMCPHub *mcpHub
var defaultMCPHubOnce sync.Once

func GetMCPHub() IMCPHub {
	defaultMCPHubOnce.Do(func() {
		defaultMCPHub = &mcpHub{
			clientMaps: make(map[string]*mcpClient),
			fnMaps:     make(map[string]*mcpClient),
		}
	})
	return defaultMCPHub
}

// ================ToolHub================
var defaultToolHub *toolHub
var defaultToolHubOnce sync.Once

func GetToolHub() IToolHub {
	defaultToolHubOnce.Do(func() {
		defaultToolHub = &toolHub{
			toolEntrys: make(map[string]ToolEntry),
			mcpSrv:     newMCPServer("tool"),
		}
	})
	return defaultToolHub
}

// ================MiddlewareHub================
var defaultMiddlewareHub *middlewareHub
var defaultMiddlewareHubOnce sync.Once

func GetMiddlewareHub() IMiddlewareHub {
	defaultMiddlewareHubOnce.Do(func() {
		defaultMiddlewareHub = &middlewareHub{
			middlewares: make(map[string]IMiddleware),
		}
	})
	return defaultMiddlewareHub
}
