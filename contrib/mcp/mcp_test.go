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

func TestExportValidation(t *testing.T) {
	r := NewRegistry()
	h := func(ctx context.Context, c *app.RequestContext) {}

	require.Error(t, Export(nil, http.MethodGet, "/a", h))
	require.Error(t, Export(r, "", "/a", h))
	require.Error(t, Export(r, http.MethodGet, "", h))
	require.Error(t, Export(r, http.MethodGet, "a", h))
	require.Error(t, Export(r, http.MethodGet, "/a", nil))
	require.NoError(t, Export(r, http.MethodGet, "/a", h))
}

func TestNew(t *testing.T) {
	engine := route.NewEngine(config.NewOptions([]config.Option{}))
	s := New(engine)

	require.NotNil(t, s)
	require.NotNil(t, s.Registry())
	require.NotNil(t, s.executor)
}
