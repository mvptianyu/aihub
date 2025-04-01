/*
@Project: aihub
@Module: core
@File : mcp_manager_test.go
*/
package core

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func Test_mcpManager_RegisterMCPService(t *testing.T) {
	mcpServerAddrs := []string{"http://localhost:8811/sse"}
	if err := GetDefaultMCPManager().RegisterMCPService(mcpServerAddrs...); err != nil {
		log.Panic(err)
	}

	tools := GetDefaultMCPManager().GetToolFunctions(mcpServerAddrs...)
	fmt.Println(tools)

	input := &ToolInputBase{}
	input.SetRawFuncName("mms_log_query_by_keyword")
	input.SetRawInput("{\"keyword\":\"sg-11134201-7rd6w-m7qad2oq19n848\"}")
	output := &Message{}
	err := GetDefaultMCPManager().ProxyMCPCall(context.Background(), input, output)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(output)
}
