/*
@Project: aihub
@Module: core
@File : config.go
*/
package core

import "os"

// AgentConfig agent配置结构
type AgentConfig struct {
	// LLM提供商配置
	Provider ProviderConfig `json:"provider" yaml:"provider"`

	MaxStoreHistory *int `json:"c,omitempty" yaml:"max_store_history,omitempty"`             // 限制总体缓存会话记忆条数
	MaxUseHistory   *int `json:"max_use_history,omitempty" yaml:"max_use_history,omitempty"` // 限制请求时使用的消息条数，避免输入token泛滥
	MaxStepQuit     *int `json:"max_step_quit,omitempty" yaml:"max_step_quit,omitempty"`     // 限制单次会话的最大执行步数，避免AI死循环

	MaxTokens        *int     `json:"max_tokens,omitempty" yaml:"max_tokens,omitempty"`               // 限制最大token数
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty" yaml:"frequency_penalty,omitempty"` // 频率惩罚[-2.0~2.0]，值越大，模型越倾向于避免重复已经生成过的词
	PresencePenalty  *float64 `json:"presence_penalty,omitempty" yaml:"presence_penalty,omitempty"`   // 存在惩罚[-2.0~2.0]，值越大，模型生成的文本中重复出现的词就越少
	Temperature      *float64 `json:"temperature,omitempty" yaml:"temperature,omitempty"`             // 温度[0.0~2.0]，值越大，模型生成的文本灵活性更高

	SystemPrompt string          `json:"system_prompt" yaml:"system_prompt"` // 系统提示词
	Tools        []*ToolFunction `json:"toolMethods" yaml:"toolMethods"`     // 用到的工具
}

func (cfg *AgentConfig) AutoFix() error {
	err := cfg.Provider.AutoFix()
	if err != nil {
		return err
	}

	if cfg.MaxStoreHistory == nil || *cfg.MaxStoreHistory > 30 {
		val := 30
		cfg.MaxStoreHistory = &val
	}
	if cfg.MaxUseHistory == nil || *cfg.MaxUseHistory > 30 {
		val := 10
		cfg.MaxUseHistory = &val
	}
	if cfg.MaxStepQuit == nil || *cfg.MaxStepQuit > 20 {
		val := 20
		cfg.MaxStepQuit = &val
	}
	if cfg.MaxTokens == nil || *cfg.MaxTokens > 4096 {
		val := 4096
		cfg.MaxTokens = &val
	}
	if cfg.FrequencyPenalty == nil || *cfg.FrequencyPenalty < -2.0 || *cfg.FrequencyPenalty > 2.0 {
		val := 0.0
		cfg.FrequencyPenalty = &val
	}
	if cfg.PresencePenalty == nil || *cfg.PresencePenalty < -2.0 || *cfg.PresencePenalty > 2.0 {
		val := 0.0
		cfg.PresencePenalty = &val
	}
	if cfg.Temperature == nil || *cfg.Temperature < 0.0 || *cfg.Temperature > 2.0 {
		val := 0.3
		cfg.Temperature = &val
	}

	return nil
}

// ProviderConfig provider配置结构
type ProviderConfig struct {
	Name      string `json:"name" yaml:"name"`
	Model     string `json:"model" yaml:"model"`
	BaseURL   string `json:"base_url" yaml:"base_url"`
	Version   string `json:"version" yaml:"version"`
	APIKey    string `json:"api_key" yaml:"api_key"`
	RateLimit int    `json:"rate_limit" yaml:"rate_limit"`
}

func (cfg *ProviderConfig) AutoFix() error {
	if cfg.Version == "" {
		cfg.Version = "v1"
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = 50
	}

	if cfg.APIKey == "" {
		// 取环境变量值
		cfg.APIKey = os.Getenv("OPENAI_API_KEY")
	}

	if cfg.Name == "" || cfg.Model == "" || cfg.APIKey == "" || cfg.BaseURL == "" {
		return ErrConfiguration
	}
	return nil
}
