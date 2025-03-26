package main

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
)

func main() {
	ctx := context.Background()

	// Create a new agent
	myAgent := aihub.NewAgentWithYamlFile("E:\\goproj\\mvptianyu\\aihub\\examples\\demo.yaml")
	myAgent.Init(Dispath)

	msg, txt, err := myAgent.Run(
		ctx,
		"深圳、香港、北京今天天气如何呢，并且根据各城市天气情况推荐一首匹配的歌名",
	)
	fmt.Println(msg, err)
	fmt.Println("=======================")
	fmt.Println(txt)
}
