package depency

import (
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/examples/depency/llms"
	"github.com/mvptianyu/aihub/examples/depency/middlewares"
	"github.com/mvptianyu/aihub/examples/depency/tools"
)

func Init() {
	InitLLMs()
	InitMCPs()
	InitTools()
	InitMiddlewares()
}

func InitLLMs() {
	_, err := aihub.GetLLMHub().SetLLMByYamlData([]byte(llms.OPENAI_GPT_3_5_TURBO))
	if err != nil {
		panic(err)
	}

	_, err = aihub.GetLLMHub().SetLLMByYamlData([]byte(llms.OPENAI_GPT_4O))
	if err != nil {
		panic(err)
	}
}

func InitMCPs() {
	err := aihub.GetMCPHub().SetClient("http://localhost:8811/sse")
	if err != nil {
		panic(err)
	}
}

func InitTools() {
	err := aihub.GetToolHub().SetTool(
		aihub.ToolEntry{
			Function:    tools.GetWeather,
			Description: "根据城市名称获取天气情况",
		},
		aihub.ToolEntry{
			Function:    tools.GetSong,
			Description: "根据天气情况获取推荐歌曲名称",
		},
	)
	if err != nil {
		panic(err)
	}
}

func InitMiddlewares() {
	err := aihub.GetMiddlewareHub().SetMiddleware(&middlewares.Approver{})
	if err != nil {
		panic(err)
	}
}
