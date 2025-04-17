package main

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/examples/depency"
)

func main() {
	depency.Init() // 初始化

	// Create a new agent
	myAgent, err := aihub.GetAgentHub().SetAgentByYamlFile("demo.yaml")
	if err != nil {
		panic(err)
		return
	}

	// run(myAgent)
	runStream(myAgent)
}

func run(myAgent aihub.IAgent) {
	rsp := myAgent.Run(
		context.Background(),
		"深圳、香港、北京今天天气如何呢，并且根据各城市天气情况推荐一首匹配的歌名",
		aihub.WithDebug(true),
	)

	fmt.Println(rsp.Err)
	fmt.Println("=======================")
	fmt.Println(rsp.Content)
}

func runStream(myAgent aihub.IAgent) {
	rsp := myAgent.RunStream(
		context.Background(),
		"深圳、香港、北京今天天气如何呢，并且根据各城市天气情况推荐一首匹配的歌名",
		aihub.WithDebug(true),
	)

	for rsp.Next() {
		data := rsp.Current()
		if data.Err != nil {
			fmt.Println(data.Err)
		}
		fmt.Printf(data.Content)
	}
	fmt.Println(rsp.Err())
	fmt.Println("\n======[Done]=======")
}
