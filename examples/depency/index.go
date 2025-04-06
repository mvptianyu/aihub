/*
@Project: aihub
@Module: depency
@File : index.go
*/
package depency

import (
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/examples/depency/middleware"
	"github.com/mvptianyu/aihub/examples/depency/tools"
)

func Init() {
	InitProviders()
	InitMCPs()
	InitTools()
	InitMiddlewares()
}

func InitProviders() {
	_, err := aihub.GetProviderHub().SetProvider(&aihub.ProviderConfig{
		Name:    "openai",
		Model:   "gpt-3.5-turbo",
		BaseURL: "https://api.openai.com",
	})
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
		aihub.ToolEntry{
			Function:    tools.QueryClickHouse,
			Description: "根据SQL语句查询核心指标数据",
		},
	)
	if err != nil {
		panic(err)
	}
}

func InitMiddlewares() {
	err := aihub.GetMiddlewareHub().SetMiddleware(&middleware.Approver{})
	if err != nil {
		panic(err)
	}
}
