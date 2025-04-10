/*
@Project: aihub
@Module: core
@File : mcp_manager_test.go
*/
package aihub

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func Test_mcpHub_SetMCP(t *testing.T) {
	mcpServerAddrs := []string{"http://localhost:8811/sse"}
	if err := GetMCPHub().SetClient(mcpServerAddrs...); err != nil {
		log.Panic(err)
	}

	tools := GetMCPHub().GetToolFunctions(mcpServerAddrs, nil)
	fmt.Println(tools)

	output := &Message{}
	err := GetMCPHub().ProxyCall(context.Background(), "mms_log_query_by_keyword", "{\"keyword\":\"sg-11134201-7rd6w-m7qad2oq19n848\"}", output)
	if err != nil {
		log.Panic(err)
	}
}
