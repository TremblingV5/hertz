/*
 * Copyright 2026 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type Server struct {
	options  *Options
	registry *Registry
	executor *Executor
}

func (s *Server) Registry() *Registry {
	return s.registry
}

func (s *Server) Export(method, path string, handler app.HandlerFunc, opts ...ToolOption) error {
	return Export(s.registry, method, path, handler, opts...)
}

// Handler returns Hertz handler for MCP endpoint.
//
// Suggested route:
//
//	h.GET("/mcp", mcpServer.Handler())
//	h.POST("/mcp", mcpServer.Handler())

// 如果请求方法为GET，返回所有工具列表

// 如果请求方法为POST，处理RPC请求
func (s *Server) Handler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		switch string(c.Method()) {
		case consts.MethodGet:
			s.handleList(c)
		case consts.MethodPost:
			s.handleRPC(ctx, c)
		default:
			c.AbortWithMsg("method not allowed", consts.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) handleList(c *app.RequestContext) {
	c.JSON(consts.StatusOK, map[string]interface{}{
		"tools": s.registry.List(),
	})
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type callToolParams struct {
	Name      string    `json:"name"`
	Arguments ToolInput `json:"arguments"`
}

type initializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]interface{} `json:"clientInfo"`
}

func (s *Server) handleRPC(ctx context.Context, c *app.RequestContext) {
	body, err := c.Body()
	if err != nil {
		s.writeSSEResponse(c, rpcResponse{
			JSONRPC: "2.0",
			Error: &rpcError{
				Code:    -32700,
				Message: err.Error(),
			},
		})
		return
	}

	var req rpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeSSEResponse(c, rpcResponse{
			JSONRPC: "2.0",
			Error: &rpcError{
				Code:    -32700,
				Message: err.Error(),
			},
		})
		return
	}

	resp := s.dispatch(ctx, req)
	s.writeSSEResponse(c, resp)
}

func (s *Server) dispatch(ctx context.Context, req rpcRequest) rpcResponse {
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		var params initializeParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &rpcError{
				Code:    -32602,
				Message: err.Error(),
			}
			return resp
		}

		resp.Result = map[string]interface{}{
			"resultType":      "complete",
			"protocolVersion": "2025-11-25",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "hertz-mcp",
				"version": "0.1.0",
			},
		}
		return resp

	case "notifications/initialized":
		// 客户端发送 initialized 通知，表示初始化完成
		// 服务器不需要回复
		return resp

	case "tools/list":
		resp.Result = map[string]interface{}{
			"resultType": "complete",
			"tools":      s.registry.List(),
		}
		return resp

	// 调用工具
	case "tools/call":
		var params callToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &rpcError{
				Code:    -32602,
				Message: err.Error(),
			}
			return resp
		}

		out, err := s.executor.Execute(ctx, params.Name, params.Arguments)
		if err != nil {
			resp.Error = &rpcError{
				Code:    -32000,
				Message: err.Error(),
			}
			return resp
		}

		resp.Result = map[string]interface{}{
			"resultType": "complete",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": mustJSON(out),
				},
			},
		}
		return resp

	default:
		resp.Error = &rpcError{
			Code:    -32601,
			Message: fmt.Sprintf("method %q not found", req.Method),
		}
		return resp
	}
}

func (s *Server) writeSSEResponse(c *app.RequestContext, resp rpcResponse) {
	c.Response.Header.Set("Content-Type", "text/event-stream")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")

	data, err := json.Marshal(resp)
	if err != nil {
		data, _ = json.Marshal(rpcResponse{
			JSONRPC: "2.0",
			Error: &rpcError{
				Code:    -32603,
				Message: err.Error(),
			},
		})
	}

	c.SetStatusCode(consts.StatusOK)
	c.SetBodyString("event: message\n")
	c.Response.AppendBodyString("data: ")
	c.Response.AppendBody(data)
	c.Response.AppendBodyString("\n\n")
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}
