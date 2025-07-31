# AIHub

AIHub 是一个强大的 AI 工具仓库，提供基础 LLM（大型语言模型）接入和 Agent 能力，帮助开发者快速构建和部署 AI 应用。

## 功能特性

- **LLM 集成**：支持接入多种大型语言模型，提供统一的接口
- **Agent 能力**：实现智能体功能，可以执行复杂的任务和工具调用
- **工具管理**：通过 ToolHub 管理和使用各种工具
- **MCP 服务**：通过 Model Context Protocol 扩展模型能力
- **中间件支持**：提供中间件机制，支持自定义处理逻辑
- **会话管理**：维护用户会话和消息历史
- **流式响应**：支持 SSE 流式响应，提供实时交互体验

## 安装

```bash
go get github.com/mvptianyu/aihub
```

## 快速开始

### 使用 LLM

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mvptianyu/aihub"
)

func main() {
    // 创建 LLM 配置
    llmConfig := &aihub.LLMConfig{
        Provider:  "openai",
        APIKey:    "your-api-key",
        ModelName: "gpt-3.5-turbo",
    }

    // 创建 LLM 实例
    llm, err := aihub.NewLLM(llmConfig)
    if err != nil {
        log.Fatalf("Failed to create LLM: %v", err)
    }

    // 创建聊天完成
    resp, err := llm.CreateChatCompletion(context.Background(), []aihub.Message{
        {
            Role:    "user",
            Content: "Hello, how are you?",
        },
    })
    if err != nil {
        log.Fatalf("Failed to create chat completion: %v", err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

### 使用 Agent 和工具

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mvptianyu/aihub"
)

func main() {
    // 创建 Agent 配置
    agentConfig := &aihub.AgentConfig{
        LLMConfig: &aihub.LLMConfig{
            Provider:  "openai",
            APIKey:    "your-api-key",
            ModelName: "gpt-3.5-turbo",
        },
        Tools: []aihub.Tool{
            {
                Name:        "calculator",
                Description: "A calculator tool",
                Function: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
                    // 实现计算逻辑
                    return nil, nil
                },
            },
        },
    }

    // 创建 Agent 实例
    agent, err := aihub.NewAgent(agentConfig)
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // 运行 Agent
    resp, err := agent.Run(context.Background(), "Calculate 2 + 2")
    if err != nil {
        log.Fatalf("Failed to run agent: %v", err)
    }

    fmt.Println(resp)
}
```

## 示例

项目包含多个示例，展示了不同的使用场景：

- **agent_sql**：SQL 智能体示例，展示如何使用 Agent 执行 SQL 查询
- **agent_with_tools**：带工具的智能体示例，展示如何为 Agent 配置和使用工具
- **llm**：LLM 使用示例，展示如何直接使用 LLM 进行对话
- **manus**：配置文件示例，展示如何使用 YAML 配置文件配置 Agent 和 AgentHub

### 使用 Manus 和 AgentHub

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mvptianyu/aihub"
)

func main() {
    // 创建 AgentHub 实例
    hub := aihub.NewAgentHub()

    // 从配置文件创建 Manus
    manus, err := aihub.NewManusFromFile("config.yaml")
    if err != nil {
        log.Fatalf("Failed to create manus: %v", err)
    }

    // 注册 Agent 到 Hub
    for _, agent := range manus.Agents {
        hub.RegisterAgent(agent.Name, agent)
    }

    // 使用 Hub 调用特定 Agent
    resp, err := hub.Run(context.Background(), "weather", "What's the weather in Beijing?")
    if err != nil {
        log.Fatalf("Failed to run agent: %v", err)
    }

    fmt.Println(resp)
}
```

## 配置

AIHub 支持通过代码或配置文件进行配置。

### 基本配置示例

```yaml
llm:
  provider: openai
  api_key: your-api-key
  model_name: gpt-3.5-turbo

tools:
  - name: calculator
    description: A calculator tool
    schema:
      type: object
      properties:
        expression:
          type: string
          description: The expression to calculate
      required:
        - expression

middleware:
  - name: approver
    config:
      auto_approve: true
```

### Manus 配置示例

Manus 允许通过 YAML 配置文件定义多个 Agent 及其工具：

```yaml
agents:
  - name: weather
    description: "Weather agent that can provide weather information"
    llm:
      provider: openai
      api_key: ${OPENAI_API_KEY}
      model_name: gpt-3.5-turbo
    tools:
      - name: get_weather
        description: "Get weather information for a location"
        schema:
          type: object
          properties:
            location:
              type: string
              description: "The location to get weather for"
          required:
            - location

  - name: song
    description: "Song agent that can provide song recommendations"
    llm:
      provider: openai
      api_key: ${OPENAI_API_KEY}
      model_name: gpt-3.5-turbo
    tools:
      - name: search_songs
        description: "Search for songs by artist or genre"
        schema:
          type: object
          properties:
            artist:
              type: string
              description: "The artist name"
            genre:
              type: string
              description: "The music genre"
```

## 核心概念

### Agent

Agent 是 AIHub 的核心组件，它封装了 LLM 的能力，并可以配置工具和中间件。Agent 可以处理用户输入，生成响应，并在需要时调用工具。

### Tool

Tool 是 Agent 可以使用的工具，它可以执行特定的任务，如计算、查询数据库、调用 API 等。Tool 由名称、描述和 JSON Schema 定义。

### Manus

Manus 是一个配置管理器，它可以从 YAML 配置文件中加载和创建多个 Agent。Manus 使得通过配置文件管理多个 Agent 变得简单。

### AgentHub

AgentHub 是一个 Agent 管理器，它可以注册和管理多个 Agent，并根据需要调用特定的 Agent。AgentHub 使得在一个应用中使用多个专门的 Agent 变得简单。

## 高级用法

### 使用 AgentHub 管理多个 Agent

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mvptianyu/aihub"
)

func main() {
    // 创建 AgentHub 实例
    hub := aihub.NewAgentHub()

    // 创建并注册第一个 Agent
    weatherAgent, err := aihub.NewAgent(&aihub.AgentConfig{
        LLMConfig: &aihub.LLMConfig{
            Provider:  "openai",
            APIKey:    "your-api-key",
            ModelName: "gpt-3.5-turbo",
        },
    })
    if err != nil {
        log.Fatalf("Failed to create weather agent: %v", err)
    }
    hub.RegisterAgent("weather", weatherAgent)

    // 创建并注册第二个 Agent
    songAgent, err := aihub.NewAgent(&aihub.AgentConfig{
        LLMConfig: &aihub.LLMConfig{
            Provider:  "openai",
            APIKey:    "your-api-key",
            ModelName: "gpt-3.5-turbo",
        },
    })
    if err != nil {
        log.Fatalf("Failed to create song agent: %v", err)
    }
    hub.RegisterAgent("song", songAgent)

    // 使用特定的 Agent
    weatherResp, err := hub.Run(context.Background(), "weather", "What's the weather in Beijing?")
    if err != nil {
        log.Fatalf("Failed to run weather agent: %v", err)
    }
    fmt.Println("Weather response:", weatherResp)

    songResp, err := hub.Run(context.Background(), "song", "Recommend me some rock songs")
    if err != nil {
        log.Fatalf("Failed to run song agent: %v", err)
    }
    fmt.Println("Song response:", songResp)
}
```

## 贡献

欢迎贡献代码、报告问题或提出新功能建议。请遵循以下步骤：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件
