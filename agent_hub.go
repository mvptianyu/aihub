/*
@Project: aihub
@Module: aihub
@File : provider_hub.go
*/
package aihub

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type agentHub struct {
	agents map[string]IAgent // agent Name => IProvider

	lock sync.RWMutex
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
		fmt.Printf("Error reading YAML file: %s => %v\n", yamlFile, err)
		return nil, err
	}
	return h.SetAgentByYamlData(yamlData)
}
