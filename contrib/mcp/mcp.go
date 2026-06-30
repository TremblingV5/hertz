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
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
)

// Export exposes a Hertz handler as an MCP tool explicitly.
//
// Export does not register an HTTP route and does not change existing Hertz behavior.
// It only stores tool metadata and the handler into the adapter registry.
func Export(r *Registry, method, path string, handler app.HandlerFunc, opts ...ToolOption) error {
	if r == nil {
		return fmt.Errorf("mcp: nil registry")
	}
	if method == "" {
		return fmt.Errorf("mcp: empty method")
	}
	if path == "" || path[0] != '/' {
		return fmt.Errorf("mcp: path must start with '/'")
	}
	if handler == nil {
		return fmt.Errorf("mcp: nil handler")
	}

	t := &Tool{
		Name:        defaultToolName(method, path),
		Method:      method,
		Path:        path,
		Description: "",
		Handler:     handler,
		InputSchema: defaultInputSchema(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return r.Register(t)
}

// New creates a MCP adapter server bound to a Hertz engine.
//
// The engine is only used for RequestContext allocation.
// The adapter does not inspect router internals.
func New(engine *route.Engine, opts ...Option) *Server {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	registry := NewRegistry()
	executor := NewExecutor(engine, registry)

	return &Server{
		options:  o,
		registry: registry,
		executor: executor,
	}
}
