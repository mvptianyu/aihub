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
	// Create a new agent
	myAgent, err := aihub.GetAgentHub().SetAgentByYamlFile("demo.yaml")
	if err != nil {
		panic(err)
		return
	}

	_, txt, _, err := myAgent.Run(
		ctx,
		"深圳、香港、北京今天天气如何呢，并且根据各城市天气情况推荐一首匹配的歌名",
		aihub.WithContext("城市参数中随机50%拼接“中国”字符串"),
		aihub.WithDebug(true),
		aihub.WithSessionID(""),
		aihub.WithSessionData(map[string]interface{}{
			"thread_id": "u7oirumAclCZMhQB-RBXX8ubvXNAhNTyzXN4gMD2QqIClneqgHpir2gz",
		}),
	)
	fmt.Println(err)
	fmt.Println("=======================")
	fmt.Println(txt)
}
