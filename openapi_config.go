package aihub

import "fmt"

// OPENAPIInfo 定义 OPENAPIConfig 规范中的 info 部分
type OPENAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// OPENAPIRequestBody 定义 OPENAPIConfig 规范中的 requestBody 部分
type OPENAPIRequestBody struct {
	Required bool                   `json:"required"`
	Content  map[string]interface{} `json:"content"`
}

// OPENAPIResponse 定义 OPENAPIConfig 规范中的 response 部分
type OPENAPIResponse struct {
	Description string `json:"description"`
}

// OPENAPIServer 定义 OPENAPIConfig 规范中的 servers 部分
type OPENAPIServer struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

// OPENAPIOperation 定义 OPENAPIConfig 规范中的 operation 部分
type OPENAPIOperation struct {
	Summary     string                     `json:"summary"`
	Description string                     `json:"description"`
	RequestBody OPENAPIRequestBody         `json:"requestBody"`
	Responses   map[string]OPENAPIResponse `json:"responses"`
	Servers     []OPENAPIServer            `json:"servers,omitempty"`
	Tags        []string                   `json:"tags,omitempty"`
}

// OPENAPIPathItem 定义 OPENAPIConfig 规范中的 pathItem 部分
type OPENAPIPathItem struct {
	Post OPENAPIOperation `json:"post"`
}

type OPENAPITag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// OPENAPIConfig 定义 OPENAPI 规范的根结构体
type OPENAPIConfig struct {
	OpenAPI string                     `json:"openapi"`
	Info    OPENAPIInfo                `json:"info"`
	Paths   map[string]OPENAPIPathItem `json:"paths"`
	Tags    []OPENAPITag               `json:"tags"`
}

// AddToolFunction 将 ToolFunction 加入OPENAPIConfig 结构体
func (cfg *OPENAPIConfig) AddToolFunction(toolFunctions []ToolFunction, server string) {
	if cfg.Paths == nil {
		cfg.Paths = make(map[string]OPENAPIPathItem)
	}

	for _, toolFunction := range toolFunctions {
		path := fmt.Sprintf("%s", toolFunction.Name)
		operation := OPENAPIOperation{
			Summary:     toolFunction.Description,
			Description: toolFunction.Description,
			RequestBody: OPENAPIRequestBody{
				Required: true,
				Content: map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": toolFunction.Parameters,
					},
				},
			},
			Responses: map[string]OPENAPIResponse{
				"200": {
					Description: "Successful response",
				},
			},
		}
		if server != "" {
			operation.Servers = []OPENAPIServer{
				{
					Url: server,
				},
			}
			operation.Tags = []string{server}
		}
		cfg.Paths[path] = OPENAPIPathItem{
			Post: operation,
		}
	}
}
