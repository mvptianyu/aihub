/*
@Project: aihub
@Module: core
@File : config.go
*/
package core

// AgentConfig agent配置结构
type AgentConfig struct {
	// LLM提供商配置
	Provider ProviderConfig `json:"provider" yaml:"provider"`

	// 最大的会话记忆条数
	MaxChatHistory int `json:"max_chat_history" yaml:"max_chat_history"`

	// 强制结束前最大步数
	MaxStopStep int `json:"max_stop_step" yaml:"max_stop_step"`

	// 系统提示词
	SystemPrompt string `json:"system_prompt" yaml:"system_prompt"`

	// 助手提示词
	AssistantPrompt string `json:"assistant_prompt" yaml:"assistant_prompt"`
}

// ProviderConfig provider配置结构
type ProviderConfig struct {
	Name    string `json:"name" yaml:"name"`
	BaseURL string `json:"base_url" yaml:"base_url"`
	Version string `json:"version" yaml:"version"`
	APIKey  string `json:"api_key" yaml:"api_key"`
}
