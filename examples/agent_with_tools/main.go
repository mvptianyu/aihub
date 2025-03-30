package main

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/core"
	"github.com/mvptianyu/aihub/examples/tools"
)

func main() {
	ctx := context.Background()

	// Create a new agent
	myAgent := aihub.NewAgentWithYamlFile("demo.yaml", &tools.Toolkits{})

	_, txt, err := myAgent.Run(
		ctx,
		"深圳、香港、北京今天天气如何呢，并且根据各城市天气情况推荐一首匹配的歌名",
		core.WithClaim("本结果由MMS AI Agent自动生成"),
		core.WithDebug(true),
		core.WithContext("城市参数中随机50%拼接“中国”字符串"),
	)
	fmt.Println(err)
	fmt.Println("=======================")
	fmt.Println(txt)

	seatalkGroup := "LMWNqAYCQVGLGi2fGYfvHw"

	core.SendSeatalkText(seatalkGroup, core.SeaTalkText{
		Content: txt,
	})
}
