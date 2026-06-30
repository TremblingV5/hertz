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
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/stretchr/testify/require"
)

func newTestExecutor(t *testing.T) (*Executor, *Registry) {
	engine := route.NewEngine(config.NewOptions([]config.Option{}))
	r := NewRegistry()
	return NewExecutor(engine, r), r
}

func TestExecutorGet(t *testing.T) {
	e, r := newTestExecutor(t)
	h := func(ctx context.Context, c *app.RequestContext) {
		id := c.Param("id")
		verbose := c.Query("verbose")
		c.JSON(http.StatusOK, map[string]interface{}{
			"id":      id,
			"verbose": verbose,
		})
	}
	require.NoError(t, r.Register(&Tool{
		Name:    "get_user",
		Method:  http.MethodGet,
		Path:    "/user/:id",
		Handler: h,
	}))

	out, err := e.Execute(context.Background(), "get_user", ToolInput{
		Path:  map[string]interface{}{"id": "123"},
		Query: map[string]interface{}{"verbose": "true"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, out.Status)

	body, ok := out.Body.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "123", body["id"])
	require.Equal(t, "true", body["verbose"])
}

func TestExecutorNotFound(t *testing.T) {
	e, _ := newTestExecutor(t)
	_, err := e.Execute(context.Background(), "missing", ToolInput{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestExecutorPostBody(t *testing.T) {
	e, r := newTestExecutor(t)
	h := func(ctx context.Context, c *app.RequestContext) {
		var req map[string]interface{}
		require.NoError(t, c.BindJSON(&req))
		c.JSON(http.StatusCreated, req)
	}
	require.NoError(t, r.Register(&Tool{
		Name:    "create_user",
		Method:  http.MethodPost,
		Path:    "/user",
		Handler: h,
	}))

	out, err := e.Execute(context.Background(), "create_user", ToolInput{
		Body: map[string]interface{}{"name": "hertz"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, out.Status)

	body, ok := out.Body.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "hertz", body["name"])
}
