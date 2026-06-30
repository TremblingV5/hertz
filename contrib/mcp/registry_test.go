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
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	require.NotNil(t, r)

	require.Error(t, r.Register(nil))
	require.Error(t, r.Register(&Tool{Name: ""}))

	tool := &Tool{
		Name:    "get_user",
		Method:  http.MethodGet,
		Path:    "/user/:id",
		Handler: func(ctx context.Context, c *app.RequestContext) {},
	}
	require.NoError(t, r.Register(tool))

	err := r.Register(tool)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate")

	got, ok := r.Get("missing")
	require.False(t, ok)
	require.Nil(t, got)

	got, ok = r.Get("get_user")
	require.True(t, ok)
	require.Equal(t, "get_user", got.Name)
	require.Equal(t, "/user/:id", got.Path)
	require.NotNil(t, got.Handler)

	list := r.List()
	require.Len(t, list, 1)
	require.Equal(t, "get_user", list[0].Name)
	require.Nil(t, list[0].Handler)
	require.Empty(t, list[0].Method)
	require.Empty(t, list[0].Path)
}
