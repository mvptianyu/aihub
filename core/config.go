/*
@Project: aihub
@Module: core
@File : config.go
*/
package core

import "os"

// AgentConfig agent配置结构
type AgentConfig struct {
	AgentRuntimeCfg

	// LLM提供商配置
	Provider ProviderConfig `json:"provider" yaml:"provider"`

	Tools []ToolFunction `json:"tools,omitempty" yaml:"tools,omitempty"` // 用到的工具
}

func (cfg *AgentConfig) AutoFix() error {
	if err := cfg.Provider.AutoFix(); err != nil {
		return err
	}

	if err := cfg.AgentRuntimeCfg.AutoFix(&cfg.Provider); err != nil {
		return err
	}

	return nil
}

// 运行时配置
type AgentRuntimeCfg struct {
	HistoryTimeout   int64   `json:"history_timeout,omitempty" yaml:"history_timeout,omitempty"`     // 历史消息缓存过期时间秒数
	MaxStoreHistory  int     `json:"max_store_history,omitempty" yaml:"max_store_history,omitempty"` // 限制总体缓存会话记忆条数
	MaxUseHistory    int     `json:"max_use_history,omitempty" yaml:"max_use_history,omitempty"`     // 限制请求时使用的消息条数，避免输入token泛滥
	MaxStepQuit      int     `json:"max_step_quit,omitempty" yaml:"max_step_quit,omitempty"`         // 限制单次会话的最大执行步数，避免AI死循环
	MaxTokens        int     `json:"max_tokens,omitempty" yaml:"max_tokens,omitempty"`               // 限制最大token数
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty" yaml:"frequency_penalty,omitempty"` // 频率惩罚[-2.0~2.0]，值越大，模型越倾向于避免重复已经生成过的词
	PresencePenalty  float64 `json:"presence_penalty,omitempty" yaml:"presence_penalty,omitempty"`   // 存在惩罚[-2.0~2.0]，值越大，模型生成的文本中重复出现的词就越少
	Temperature      float64 `json:"temperature,omitempty" yaml:"temperature,omitempty"`             // 温度[0.0~2.0]，值越大，模型生成的文本灵活性更高

	SystemPrompt string `json:"system_prompt,omitempty" yaml:"system_prompt,omitempty"` // 系统提示词
	StopWords    string `json:"stop_words,omitempty" yaml:"stop_words,omitempty"`       // 结束退出词
	Claim        string `json:"claim,omitempty" yaml:"claim,omitempty"`                 // 宣称文案，例如：本次返回由xxx提供
	Debug        bool   `json:"debug,omitempty" yaml:"debug,omitempty"`                 // debug输出标志，开启则输出具体工具调用过程信息
}

func (cfg *AgentRuntimeCfg) AutoFix(providerCfg *ProviderConfig) error {
	if cfg.HistoryTimeout <= 0 || cfg.HistoryTimeout > 7*24*60*60 {
		cfg.HistoryTimeout = 7 * 24 * 60 * 60
	}
	if cfg.MaxStoreHistory <= 0 || cfg.MaxStoreHistory > 50 {
		cfg.MaxStoreHistory = 50
	}
	if cfg.MaxUseHistory <= 0 || cfg.MaxUseHistory > 20 {
		cfg.MaxUseHistory = 20
	}
	if cfg.MaxStepQuit <= 0 || cfg.MaxStepQuit > 20 {
		cfg.MaxStepQuit = 20
	}
	if cfg.MaxTokens <= 0 || cfg.MaxTokens > providerCfg.MaxTokens {
		cfg.MaxTokens = providerCfg.MaxTokens
	}
	if cfg.FrequencyPenalty < -2.0 || cfg.FrequencyPenalty > 2.0 {
		cfg.FrequencyPenalty = 0.0
	}
	if cfg.PresencePenalty < -2.0 || cfg.PresencePenalty > 2.0 {
		cfg.PresencePenalty = 0.0
	}
	if cfg.Temperature < 0.0 || cfg.Temperature > 2.0 {
		cfg.Temperature = 0.3
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
	MaxTokens int    `json:"max_tokens" yaml:"max_tokens"` // 模型本身限制的最大token数
	RateLimit int    `json:"rate_limit" yaml:"rate_limit"`
}

func (cfg *ProviderConfig) AutoFix() error {
	if cfg.Version == "" {
		cfg.Version = "v1"
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = 100
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096 // 默认为OPENAPI限制最大数
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
