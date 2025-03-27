/*
@Project: aihub
@Module: aihub
@File : index.go
*/
package aihub

import (
	"github.com/mvptianyu/aihub/core"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
)

// NewAgent creates a new agent with the given provider
func NewAgent(cfg *core.AgentConfig, router core.ToolFuncRouter) core.IAgent {
	return core.NewAgent(cfg, router)
}

// NewAgentWithYaml 从配置读取
func NewAgentWithYamlData(yamlData []byte, router core.ToolFuncRouter) core.IAgent {
	cfg := &core.AgentConfig{}
	if err := yaml.Unmarshal(yamlData, cfg); err != nil {
		log.Fatalf("Error Unmarshal YAML data: %s => %v\n", string(yamlData), err)
		return nil
	}

	return core.NewAgent(cfg, router)
}

// NewAgentWithYamlFile 从配置文件读取
func NewAgentWithYamlFile(yamlFile string, router core.ToolFuncRouter) core.IAgent {
	// 读取 YAML 文件内容
	yamlData, err := os.ReadFile(filepath.Clean(yamlFile))
	if err != nil {
		log.Fatalf("Error reading YAML file: %s => %v\n", yamlFile, err)
		return nil
	}

	return NewAgentWithYamlData(yamlData, router)
}

// NewProvider creates a new provider
func NewProvider(cfg *core.ProviderConfig) core.IProvider {
	return core.NewProvider(cfg)
}
