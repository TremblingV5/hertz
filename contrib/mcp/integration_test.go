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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/stretchr/testify/require"
)

func waitServerReady(t *testing.T, addr string) {
	deadline := time.Now().Add(5 * time.Second)
	url := "http://" + addr + "/v1/user/health"
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server not ready")
}

func TestIntegrationEndToEnd(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	h := server.New(
		server.WithListener(ln),
		server.WithTransport(standard.NewTransporter),
	)

	getUser := func(ctx context.Context, c *app.RequestContext) {
		id := c.Param("id")
		verbose := c.Query("verbose")
		c.JSON(http.StatusOK, map[string]interface{}{
			"id":      id,
			"verbose": verbose,
		})
	}
	h.GET("/v1/user/:id", getUser)

	mcpServer := New(h.Engine)
	require.NoError(t, mcpServer.Export(
		http.MethodGet,
		"/v1/user/:id",
		getUser,
		WithDescription("Get user by id"),
	))

	h.GET("/mcp", mcpServer.Handler())
	h.POST("/mcp", mcpServer.Handler())

	go h.Spin()
	defer h.Close()

	waitServerReady(t, ln.Addr().String())

	base := "http://" + ln.Addr().String()

	// 1. GET /mcp -> tools/list
	resp, err := http.Get(base + "/mcp")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResp))
	require.NoError(t, resp.Body.Close())

	tools := listResp["tools"].([]interface{})
	require.Len(t, tools, 1)
	tool := tools[0].(map[string]interface{})
	require.Equal(t, "get_v1_user_id", tool["name"])
	require.Equal(t, "Get user by id", tool["description"])

	// 2. POST /mcp -> tools/call
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "get_v1_user_id",
			"arguments": map[string]interface{}{
				"path":  map[string]interface{}{"id": "123"},
				"query": map[string]interface{}{"verbose": "true"},
			},
		},
	}
	b, err := json.Marshal(reqBody)
	require.NoError(t, err)

	postResp, err := http.Post(base+"/mcp", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, postResp.StatusCode)

	var dataLine string
	scanner := bufio.NewScanner(postResp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			dataLine = strings.TrimPrefix(line, "data: ")
			break
		}
	}
	require.NoError(t, scanner.Err())
	require.NotEmpty(t, dataLine)
	require.NoError(t, postResp.Body.Close())

	var rpcResp rpcResponse
	require.NoError(t, json.Unmarshal([]byte(dataLine), &rpcResp))
	require.Nil(t, rpcResp.Error)
	require.Equal(t, 1.0, rpcResp.ID)

	content := rpcResp.Result.(map[string]interface{})["content"].([]interface{})
	require.Len(t, content, 1)
	text := content[0].(map[string]interface{})["text"].(string)

	var toolResp ToolResponse
	require.NoError(t, json.Unmarshal([]byte(text), &toolResp))
	require.Equal(t, http.StatusOK, toolResp.Status)

	bodyMap, ok := toolResp.Body.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "123", bodyMap["id"])
	require.Equal(t, "true", bodyMap["verbose"])
}
