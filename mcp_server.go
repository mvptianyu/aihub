package aihub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"net/http"
	"sync"
)

const DefaultHTTPSessionID = "DEFAULT_HTTP_SESSION_ID"

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

	// HTTP直调兼容
	if r.URL.Path != "" && r.URL.Path == s.GetMessagePath() && r.URL.Query().Get("sessionId") == DefaultHTTPSessionID {
		s.directHandleMessage(w, r)
		return
	}

	s.sseSrv.ServeHTTP(w, r)
}

func (s *mcpServer) directHandleMessage(w http.ResponseWriter, r *http.Request) {
	// Parse message as raw JSON
	var rawMessage json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawMessage); err != nil {
		errRsp := mcp.JSONRPCError{
			JSONRPC: mcp.JSONRPC_VERSION,
			ID:      nil,
			Error: struct {
				Code    int         `json:"code"`
				Message string      `json:"message"`
				Data    interface{} `json:"data,omitempty"`
			}{
				Code:    mcp.PARSE_ERROR,
				Message: "Parse error",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errRsp)
		return
	}

	// Process message through MCPServer
	response := s.mcpSrv.HandleMessage(r.Context(), rawMessage)
	// Only send response if there is one (not for notifications)
	if response != nil {
		// Send HTTP response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		// For notifications, just send 202 Accepted with no body
		w.WriteHeader(http.StatusAccepted)
	}
}
