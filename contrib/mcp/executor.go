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
	"net/url"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/cloudwego/hertz/pkg/route/param"
)

type Executor struct {
	engine   *route.Engine
	registry *Registry
}

type ToolInput struct {
	Path  map[string]interface{} `json:"path,omitempty"`
	Query map[string]interface{} `json:"query,omitempty"`
	Body  interface{}            `json:"body,omitempty"`
}

type ToolResponse struct {
	Status int                 `json:"status"`
	Header map[string][]string `json:"header,omitempty"`
	Body   interface{}         `json:"body,omitempty"`
}

func NewExecutor(engine *route.Engine, registry *Registry) *Executor {
	return &Executor{
		engine:   engine,
		registry: registry,
	}
}

func (e *Executor) Execute(ctx context.Context, toolName string, input ToolInput) (*ToolResponse, error) {
	if e.engine == nil {
		return nil, fmt.Errorf("mcp: nil hertz engine")
	}

	tool, ok := e.registry.Get(toolName)
	if !ok {
		return nil, fmt.Errorf("mcp: tool %q not found", toolName)
	}

	rc := e.acquireContext()
	defer e.releaseContext(rc)

	requestPath, params := buildPathAndParams(tool.Path, input.Path)

	rc.Params = append(rc.Params[:0], params...)
	rc.SetFullPath(tool.Path)
	rc.SetIndex(-1)

	rc.Request.SetMethod(tool.Method)

	query := encodeQuery(input.Query)
	if query != "" {
		rc.Request.SetRequestURI(requestPath + "?" + query)
		rc.Request.SetQueryString(query)
	} else {
		rc.Request.SetRequestURI(requestPath)
	}
	rc.Request.URI().SetPath(requestPath)

	if input.Body != nil && !strings.EqualFold(tool.Method, consts.MethodGet) && !strings.EqualFold(tool.Method, consts.MethodHead) {
		body, err := json.Marshal(input.Body)
		if err != nil {
			return nil, fmt.Errorf("mcp: marshal request body: %w", err)
		}
		rc.Request.SetBody(body)
		rc.Request.Header.SetContentTypeBytes([]byte("application/json"))
		rc.Request.Header.SetContentLength(len(body))
	}

	tool.Handler(ctx, rc)

	return parseResponse(rc), nil
}

func (e *Executor) acquireContext() *app.RequestContext {
	v := e.engine.GetCtxPool().Get()
	if v == nil {
		return e.engine.NewContext()
	}
	return v.(*app.RequestContext)
}

func (e *Executor) releaseContext(ctx *app.RequestContext) {
	if ctx == nil {
		return
	}
	if ctx.IsExiled() {
		return
	}
	ctx.Reset()
	e.engine.GetCtxPool().Put(ctx)
}

func encodeQuery(values map[string]interface{}) string {
	if len(values) == 0 {
		return ""
	}

	q := url.Values{}
	for k, v := range values {
		switch vv := v.(type) {
		case []interface{}:
			for _, item := range vv {
				q.Add(k, fmt.Sprint(item))
			}
		case []string:
			for _, item := range vv {
				q.Add(k, item)
			}
		default:
			q.Set(k, fmt.Sprint(v))
		}
	}
	return q.Encode()
}

func buildPathAndParams(pattern string, values map[string]interface{}) (string, param.Params) {
	if values == nil {
		values = map[string]interface{}{}
	}

	parts := strings.Split(pattern, "/")
	params := make(param.Params, 0)

	for i, part := range parts {
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, ":") {
			key := strings.TrimPrefix(part, ":")
			value := fmt.Sprint(values[key])
			parts[i] = url.PathEscape(value)
			params = append(params, param.Param{
				Key:   key,
				Value: value,
			})
			continue
		}

		if strings.HasPrefix(part, "*") {
			key := strings.TrimPrefix(part, "*")
			value := fmt.Sprint(values[key])
			parts[i] = strings.TrimPrefix(value, "/")
			params = append(params, param.Param{
				Key:   key,
				Value: value,
			})
		}
	}

	path := strings.Join(parts, "/")
	if path == "" || path[0] != '/' {
		path = "/" + path
	}
	return path, params
}

func parseResponse(ctx *app.RequestContext) *ToolResponse {
	headers := make(map[string][]string)

	ctx.Response.Header.VisitAll(func(k, v []byte) {
		key := string(k)
		headers[key] = append(headers[key], string(v))
	})

	bodyBytes := ctx.Response.Body()
	var body interface{}

	contentType := strings.ToLower(ctx.Response.Header.Get("Content-Type"))
	if len(bodyBytes) > 0 && strings.Contains(contentType, "json") {
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			body = string(bodyBytes)
		}
	} else if len(bodyBytes) > 0 {
		body = string(bodyBytes)
	}

	return &ToolResponse{
		Status: ctx.Response.StatusCode(),
		Header: headers,
		Body:   body,
	}
}
