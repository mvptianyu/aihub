/*
@Project: aihub
@Module: depency
@File : index.go
*/
package depency

import (
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/examples/depency/middlewares"
	"github.com/mvptianyu/aihub/examples/depency/providers"
	"github.com/mvptianyu/aihub/examples/depency/tools"
)

func Init() {
	InitProviders()
	InitMCPs()
	InitTools()
	InitMiddlewares()
}

func InitProviders() {
	_, err := aihub.GetProviderHub().SetProviderByYamlData([]byte(providers.OPENAI_CONFIG))
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
