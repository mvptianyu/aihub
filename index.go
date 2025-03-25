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
)

// NewAgent creates a new agent with the given provider
func NewAgent(cfg *core.AgentConfig) core.IAgent {
	return core.NewAgent(cfg)
}

// NewAgentWithYaml 从配置读取
func NewAgentWithYamlData(yamlData []byte) core.IAgent {
	cfg := &core.AgentConfig{}
	if err := yaml.Unmarshal(yamlData, cfg); err != nil {
		log.Fatalf("Error Unmarshal YAML data: %s => %v\n", string(yamlData), err)
		return nil
	}

	return core.NewAgent(cfg)
}

// NewAgentWithYamlFile 从配置文件读取
func NewAgentWithYamlFile(yamlFile string) core.IAgent {
	// 读取 YAML 文件内容
	yamlData, err := os.ReadFile(yamlFile)
	if err != nil {
		log.Fatalf("Error reading YAML file: %s => %v\n", yamlFile, err)
		return nil
	}

	return NewAgentWithYamlData(yamlData)
}

// NewProvider creates a new provider
func NewProvider(cfg *core.ProviderConfig) core.IProvider {
	return core.NewProvider(cfg)
}
