package main

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/examples/tools"
)

func main() {
	ctx := context.Background()

	// Create a new agent
	myAgent := aihub.NewAgentWithYamlFile("demo.yaml", &tools.Toolkits{})
	myAgent.RegisterMiddleware(&tools.Approver{})

	_, txt, err := myAgent.Run(
		ctx,
		"深圳、香港、北京今天天气如何呢，并且根据各城市天气情况推荐一首匹配的歌名，然后帮我查一下sg-11134201-7rd6w-m7qad2oq19n848的日志",
		aihub.WithContext("城市参数中随机50%拼接“中国”字符串"),
		aihub.WithDebug(true),
		aihub.WithSessionData(map[string]interface{}{
			"thread_id": "u7oirumAclCZMhQB-RBXX8ubvXNAhNTyzXN4gMD2QqIClneqgHpir2gz",
		}),
	)
	fmt.Println(err)
	fmt.Println("=======================")
	fmt.Println(txt)

	tools.SendSeatalkText(tools.SeatalkGroup, tools.SeaTalkText{
		Content: txt,
	})
}
