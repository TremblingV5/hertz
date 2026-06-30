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

type Options struct {
	Endpoint string
}

type Option func(*Options)

func defaultOptions() *Options {
	return &Options{
		Endpoint: "/mcp",
	}
}

// WithEndpoint customizes MCP endpoint path.
func WithEndpoint(path string) Option {
	return func(o *Options) {
		if path != "" {
			o.Endpoint = path
		}
	}
}

type ToolOption func(*Tool)

// WithName customizes MCP tool name.
func WithName(name string) ToolOption {
	return func(t *Tool) {
		if name != "" {
			t.Name = name
		}
	}
}

// WithDescription sets MCP tool description.
func WithDescription(desc string) ToolOption {
	return func(t *Tool) {
		t.Description = desc
	}
}

// WithInputSchema overrides the default input schema.
func WithInputSchema(schema map[string]interface{}) ToolOption {
	return func(t *Tool) {
		if schema != nil {
			t.InputSchema = schema
		}
	}
}
