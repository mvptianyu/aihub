package main

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/examples/depency"
)

func main() {
	depency.Init() // 初始化

	ctx := context.Background()
	aihub.GetAgentHub().SetAgentByYamlFile("weather.yaml")
	aihub.GetAgentHub().SetAgentByYamlFile("song.yaml")

	_, txt, _, err := aihub.GetManus().Run(ctx, "深圳、香港、北京今天天气如何呢，并且根据各城市天气情况推荐一首匹配的歌名",
		aihub.WithAgents([]string{"weather", "song"}),
		// aihub.WithSystemPrompt(""),
	)
	fmt.Println(err)
	fmt.Println("=======================")
	fmt.Println(txt)
}
