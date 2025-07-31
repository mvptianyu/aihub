package aihub

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	"net/http"
	"sync"
)

type mcpServer struct {
	mcpSrv *server.MCPServer
	sseSrv *server.SSEServer
	name   string

	lock sync.RWMutex
}

func newMCPServer(name string) IMCPServer {
	ret := &mcpServer{
		mcpSrv: server.NewMCPServer(
			fmt.Sprintf("aihub-mcp-%s", name),
			"1.0.0",
			server.WithToolCapabilities(true),
		),
		name: name,
	}

	ret.sseSrv = server.NewSSEServer(ret.mcpSrv,
		server.WithSSEEndpoint(fmt.Sprintf("/sse_%s", name)),
		server.WithMessageEndpoint(fmt.Sprintf("/message_%s", name)),
	)

	return ret
}

func (s *mcpServer) Start(listenAddr string) error {
	if s.sseSrv == nil {
		return fmt.Errorf("sseSrv is nil")
	}

	return s.sseSrv.Start(listenAddr)
}

func (s *mcpServer) Shutdown(ctx context.Context) error {
	if s.sseSrv == nil {
		return fmt.Errorf("sseSrv is nil")
	}

	return s.sseSrv.Shutdown(ctx)
}

func (s *mcpServer) GetSSEPath() string {
	if s.sseSrv == nil {
		return ""
	}
	return s.sseSrv.CompleteSsePath()
}

func (s *mcpServer) GetMessagePath() string {
	if s.sseSrv == nil {
		return ""
	}
	return s.sseSrv.CompleteMessagePath()
}

func (s *mcpServer) AddTools(tools ...server.ServerTool) {
	if s.mcpSrv == nil {
		return
	}
	s.mcpSrv.AddTools(tools...)
}

func (s *mcpServer) DelTools(names ...string) {
	if s.mcpSrv == nil {
		return
	}
	s.mcpSrv.DeleteTools(names...)
}

func (s *mcpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.sseSrv == nil {
		return
	}
	s.sseSrv.ServeHTTP(w, r)
}
