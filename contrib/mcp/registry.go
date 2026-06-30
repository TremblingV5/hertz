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
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
)

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Method      string                 `json:"-"`
	Path        string                 `json:"-"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
	Handler     app.HandlerFunc        `json:"-"`
}

type Registry struct {
	mu    sync.RWMutex
	tools map[string]*Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]*Tool),
	}
}

func (r *Registry) Register(t *Tool) error {
	if t == nil {
		return fmt.Errorf("mcp: nil tool")
	}
	if t.Name == "" {
		return fmt.Errorf("mcp: empty tool name")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tools[t.Name]; ok {
		return fmt.Errorf("mcp: duplicate tool name %q", t.Name)
	}

	cp := *t
	r.tools[t.Name] = &cp
	return nil
}

func (r *Registry) Get(name string) (*Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tools[name]
	if !ok {
		return nil, false
	}

	cp := *t
	return &cp, true
}

func (r *Registry) List() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*Tool, 0, len(r.tools))
	for _, t := range r.tools {
		cp := *t
		cp.Handler = nil
		cp.Method = ""
		cp.Path = ""
		out = append(out, &cp)
	}
	return out
}

func defaultInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "object",
				"description": "URL path parameters",
			},
			"query": map[string]interface{}{
				"type":        "object",
				"description": "Query string parameters",
			},
			"body": map[string]interface{}{
				"type":        "object",
				"description": "JSON request body",
			},
		},
	}
}
