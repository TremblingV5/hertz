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
	"net/http"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) *Server {
	engine := route.NewEngine(config.NewOptions([]config.Option{}))
	s := New(engine)
	h := func(ctx context.Context, c *app.RequestContext) {
		id := c.Param("id")
		c.JSON(http.StatusOK, map[string]interface{}{"id": id})
	}
	require.NoError(t, s.Export(http.MethodGet, "/items/:id", h, WithDescription("Get item")))
	return s
}

func TestServerHandlerList(t *testing.T) {
	s := newTestServer(t)
	c := app.NewContext(16)
	c.Request.SetMethod(consts.MethodGet)
	c.Request.SetRequestURI("/mcp")

	s.Handler()(context.Background(), c)

	require.Equal(t, consts.StatusOK, c.Response.StatusCode())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(c.Response.Body(), &resp))

	tools, ok := resp["tools"].([]interface{})
	require.True(t, ok)
	require.Len(t, tools, 1)

	tool := tools[0].(map[string]interface{})
	require.Equal(t, "get_items_id", tool["name"])
	require.Equal(t, "Get item", tool["description"])
}

func TestServerHandlerCall(t *testing.T) {
	s := newTestServer(t)
	c := app.NewContext(16)
	c.Request.SetMethod(consts.MethodPost)
	c.Request.SetRequestURI("/mcp")

	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "get_items_id",
			"arguments": map[string]interface{}{
				"path": map[string]interface{}{"id": "42"},
			},
		},
	}
	b, err := json.Marshal(body)
	require.NoError(t, err)
	c.Request.SetBody(b)
	c.Request.Header.SetContentLength(len(b))
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))

	s.Handler()(context.Background(), c)

	require.Equal(t, consts.StatusOK, c.Response.StatusCode())
	require.Equal(t, "text/event-stream", c.Response.Header.Get("Content-Type"))

	raw := string(c.Response.Body())
	require.True(t, strings.HasPrefix(raw, "event: message\n"))

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	require.Len(t, lines, 2)

	var data rpcResponse
	dataLine := strings.TrimPrefix(lines[1], "data: ")
	require.NoError(t, json.Unmarshal([]byte(dataLine), &data))
	require.Equal(t, 1.0, data.ID)
	require.Nil(t, data.Error)
	require.NotNil(t, data.Result)

	result, ok := data.Result.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "complete", result["resultType"])
}

func TestServerHandlerInitialize(t *testing.T) {
	s := newTestServer(t)
	c := app.NewContext(16)
	c.Request.SetMethod(consts.MethodPost)
	c.Request.SetRequestURI("/mcp")

	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2025-11-25",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}
	b, err := json.Marshal(body)
	require.NoError(t, err)
	c.Request.SetBody(b)
	c.Request.Header.SetContentLength(len(b))
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))

	s.Handler()(context.Background(), c)

	require.Equal(t, consts.StatusOK, c.Response.StatusCode())

	raw := string(c.Response.Body())
	require.True(t, strings.HasPrefix(raw, "event: message\n"))

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	require.Len(t, lines, 2)

	var data rpcResponse
	dataLine := strings.TrimPrefix(lines[1], "data: ")
	require.NoError(t, json.Unmarshal([]byte(dataLine), &data))
	require.Equal(t, 1.0, data.ID)
	require.Nil(t, data.Error)
	require.NotNil(t, data.Result)

	result, ok := data.Result.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "complete", result["resultType"])
	require.Equal(t, "2025-11-25", result["protocolVersion"])

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "hertz-mcp", serverInfo["name"])
	require.Equal(t, "0.1.0", serverInfo["version"])

	caps, ok := result["capabilities"].(map[string]interface{})
	require.True(t, ok)
	tools, ok := caps["tools"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, true, tools["listChanged"])
}

func TestServerHandlerInitializedNotification(t *testing.T) {
	s := newTestServer(t)
	c := app.NewContext(16)
	c.Request.SetMethod(consts.MethodPost)
	c.Request.SetRequestURI("/mcp")

	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	b, err := json.Marshal(body)
	require.NoError(t, err)
	c.Request.SetBody(b)
	c.Request.Header.SetContentLength(len(b))
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))

	s.Handler()(context.Background(), c)

	require.Equal(t, consts.StatusOK, c.Response.StatusCode())

	raw := string(c.Response.Body())
	require.True(t, strings.HasPrefix(raw, "event: message\n"))

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	require.Len(t, lines, 2)

	var data rpcResponse
	dataLine := strings.TrimPrefix(lines[1], "data: ")
	require.NoError(t, json.Unmarshal([]byte(dataLine), &data))
	// 通知不需要回复，result 和 error 都应为 nil
	require.Nil(t, data.Error)
	require.Nil(t, data.Result)
}

func TestServerHandlerMethodNotAllowed(t *testing.T) {
	s := newTestServer(t)
	c := app.NewContext(16)
	c.Request.SetMethod(consts.MethodPut)
	c.Request.SetRequestURI("/mcp")

	s.Handler()(context.Background(), c)

	require.Equal(t, consts.StatusMethodNotAllowed, c.Response.StatusCode())
}

func TestServerHandlerInvalidJSON(t *testing.T) {
	s := newTestServer(t)
	c := app.NewContext(16)
	c.Request.SetMethod(consts.MethodPost)
	c.Request.SetRequestURI("/mcp")
	c.Request.SetBody([]byte("not json"))

	s.Handler()(context.Background(), c)

	require.Equal(t, consts.StatusOK, c.Response.StatusCode())

	raw := string(c.Response.Body())
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	var data rpcResponse
	require.NoError(t, json.Unmarshal([]byte(strings.TrimPrefix(lines[1], "data: ")), &data))
	require.NotNil(t, data.Error)
	require.Equal(t, -32700, data.Error.Code)
}
