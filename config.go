package aihub

import (
	"fmt"
	"os"
	"strings"
)

// BriefInfo 公共基础简介
type BriefInfo struct {
	Name        string `json:"name" yaml:"name"`               // 名称
	Description string `json:"description" yaml:"description"` // 描述
}

// AgentConfig agent配置结构
type AgentConfig struct {
	BriefInfo       `yaml:",inline"` // yaml解析inline结构
	AgentRuntimeCfg `yaml:",inline"` // yaml解析inline结构

	Tools       []string               `json:"tools,omitempty" yaml:"tools,omitempty"`               // 用到的工具名
	Mcps        []string               `json:"mcps,omitempty" yaml:"mcps,omitempty"`                 // 用到的MCP服务
	Middlewares []string               `json:"middlewares,omitempty" yaml:"middlewares,omitempty"`   // 用到的Middleware
	SessionData map[string]interface{} `json:"session_data,omitempty" yaml:"session_data,omitempty"` // 用到的Session数据
}

func (cfg *AgentConfig) AutoFix() error {
	if err := cfg.AgentRuntimeCfg.AutoFix(); err != nil {
		return err
	}

	if cfg.Tools == nil {
		cfg.Tools = make([]string, 0)
	}
	if cfg.Mcps == nil {
		cfg.Mcps = make([]string, 0)
	}
	if cfg.Middlewares == nil {
		cfg.Middlewares = make([]string, 0)
	}
	if cfg.SessionData == nil {
		cfg.SessionData = make(map[string]interface{})
	}

	return nil
}

// AgentRuntimeCfg 运行时配置
type AgentRuntimeCfg struct {
	MemoryTimeout    int64   `json:"memory_timeout,omitempty" yaml:"memory_timeout,omitempty"`       // 历史消息缓存过期时间秒数
	MaxStoreMemory   int     `json:"max_store_memory,omitempty" yaml:"max_store_memory,omitempty"`   // 限制总体缓存会话记忆条数
	MaxUseMemory     int     `json:"max_use_memory,omitempty" yaml:"max_use_memory,omitempty"`       // 限制请求时使用的消息条数，避免输入token泛滥
	MaxStepQuit      int     `json:"max_step_quit,omitempty" yaml:"max_step_quit,omitempty"`         // 限制单次会话的最大执行步数，避免AI死循环
	MaxTokens        int     `json:"max_tokens,omitempty" yaml:"max_tokens,omitempty"`               // 限制最大token数
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty" yaml:"frequency_penalty,omitempty"` // 频率惩罚[-2.0~2.0]，值越大，模型越倾向于避免重复已经生成过的词
	PresencePenalty  float64 `json:"presence_penalty,omitempty" yaml:"presence_penalty,omitempty"`   // 存在惩罚[-2.0~2.0]，值越大，模型生成的文本中重复出现的词就越少
	Temperature      float64 `json:"temperature,omitempty" yaml:"temperature,omitempty"`             // 温度[0.0~2.0]，值越大，模型生成的文本灵活性更高

	LLM          string `json:"llm,omitempty" yaml:"llm,omitempty"`                     // LLM提供商配置
	SystemPrompt string `json:"system_prompt,omitempty" yaml:"system_prompt,omitempty"` // 系统提示词
	StopWords    string `json:"stop_words,omitempty" yaml:"stop_words,omitempty"`       // 结束退出词
	RunTimeout   int64  `json:"run_timeout,omitempty" yaml:"run_timeout,omitempty"`     // 执行超时秒数
	Claim        string `json:"claim,omitempty" yaml:"claim,omitempty"`                 // 宣称文案，例如：本次返回由xxx提供
	Debug        bool   `json:"debug,omitempty" yaml:"debug,omitempty"`                 // debug输出标志，开启则输出具体工具调用过程信息
}

func (cfg *AgentRuntimeCfg) AutoFix() error {
	if cfg.MemoryTimeout <= 0 || cfg.MemoryTimeout > 7*24*60*60 {
		cfg.MemoryTimeout = 7 * 24 * 60 * 60
	}
	if cfg.MaxStoreMemory <= 0 || cfg.MaxStoreMemory > 50 {
		cfg.MaxStoreMemory = 50
	}
	if cfg.MaxUseMemory <= 0 || cfg.MaxUseMemory > 20 {
		cfg.MaxUseMemory = 20
	}
	if cfg.MaxStepQuit <= 0 || cfg.MaxStepQuit > 20 {
		cfg.MaxStepQuit = 20
	}
	if cfg.MaxTokens <= 0 || cfg.MaxTokens > 4096 {
		cfg.MaxTokens = 4096
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

	if cfg.RunTimeout <= 0 || cfg.RunTimeout > 60*60 {
		cfg.RunTimeout = 60 * 60
	}

	if cfg.LLM == "" {
		return ErrConfiguration
	}

	return nil
}

type LLMType int

const (
	LLMType_Base   LLMType = iota // 基础模型，例GPT-3.5-turbo、LLM3.2等
	LLMType_Reason                // 推理模型，例GPT-4o、Deepseek R1等
	LLMType_Vision                // 视觉模型，例GPT-4o等
)

// LLMConfig provider配置结构
type LLMConfig struct {
	BriefInfo `yaml:",inline"` // yaml解析inline结构

	ModelType LLMType `json:"model_type" yaml:"model_type"`
	Provider  string  `json:"provider" yaml:"provider"` // 提供商名称，例如openai
	BaseURL   string  `json:"base_url" yaml:"base_url"`
	Version   string  `json:"version" yaml:"version"`
	APIKey    string  `json:"api_key" yaml:"api_key"`
	MaxTokens int     `json:"max_tokens" yaml:"max_tokens"` // 模型本身限制的最大token数
	RateLimit int     `json:"rate_limit" yaml:"rate_limit"`
}

func (cfg *LLMConfig) AutoFix() error {
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
		cfg.APIKey = os.Getenv(fmt.Sprintf("%s_API_KEY", strings.ToUpper(cfg.Provider)))
	}

	if cfg.Name == "" || cfg.Provider == "" || cfg.APIKey == "" || cfg.BaseURL == "" {
		return ErrConfiguration
	}
	return nil
}
