/*
@Project: aihub
@Module: aihub
@File : index.go
*/
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
		}
	})
	return defaultAgentHub
}

// ================ProviderHub================
var defaultProviderHub *providerHub
var defaultProviderHubOnce sync.Once

func GetProviderHub() IProviderHub {
	defaultProviderHubOnce.Do(func() {
		defaultProviderHub = &providerHub{
			providers: make(map[string]IProvider),
		}
	})
	return defaultProviderHub
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
