/*
 * Copyright 2026 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package main

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/contrib/mcp"
)

func main() {
	h := server.Default(server.WithHostPorts(":8080"))

	getUser := func(ctx context.Context, c *app.RequestContext) {
		id := c.Param("id")
		verbose := c.Query("verbose")

		c.JSON(http.StatusOK, map[string]interface{}{
			"id":      id,
			"name":    "hertz",
			"verbose": verbose,
		})
	}

	// Original Hertz route remains unchanged.
	h.GET("/v1/user/test", getUser)

	mcpServer := mcp.New(h.Engine)

	// Explicit MCP export.
	if err := mcpServer.Export(
		http.MethodGet,
		"/v1/user/test",
		getUser,
		mcp.WithDescription(" test"),
	); err != nil {
		panic(err)
	}

	h.GET("/mcp", mcpServer.Handler())
	h.POST("/mcp", mcpServer.Handler())

	h.Spin()
}
