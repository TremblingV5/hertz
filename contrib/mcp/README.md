# mcp

CloudWeGo Hertz 的 MCP（Model Context Protocol）适配器。将已有的 Hertz 处理器暴露为 MCP 工具，无需修改现有路由，也不使用反射。当前处于概念验证阶段，使用固定schema，后续可以支持参数解析，传入body结构体等控制schema。
## 安装

```bash
go get github.com/cloudwego/hertz/contrib/mcp
```

## 快速开始

```go
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

	// 定义处理器
	getUser := func(ctx context.Context, c *app.RequestContext) {
		id := c.Param("id")
		c.JSON(http.StatusOK, map[string]interface{}{"id": id})
	}

	// 原有 Hertz 路由（保持不变）
	h.GET("/v1/user/:id", getUser)

	// 创建 MCP 服务器并导出工具
	mcpServer := mcp.New(h.Engine)
	mcpServer.Export(
		http.MethodGet,
		"/v1/user/:id",
		getUser,
		mcp.WithDescription("根据 id 获取用户"),
	)

	// 注册 MCP 端点
	h.GET("/mcp", mcpServer.Handler())
	h.POST("/mcp", mcpServer.Handler())

	h.Spin()
}
```

## MCP 生命

服务器支持完整的 MCP 协议生命周期：

1. **`initialize`** — 客户端发送握手请求，服务器返回协议版本和能力
2. **`notifications/initialized`** — 客户端通知初始化完成
3. **`tools/list`** — 列出所有已注册的工具
4. **`tools/call`** — 通过名称和参数调用工具

## API

### `New(engine *route.Engine, opts ...Option) *Server`

创建绑定到 Hertz 引擎的 MCP 适配器服务器。

### `Export(r *Registry, method, path string, handler app.HandlerFunc, opts ...ToolOption) error`

将 Hertz 处理器暴露为 MCP 工具。不会注册 HTTP 路由。

### `Server.Handler() app.HandlerFunc`

返回 MCP 端点的 Hertz 处理器。支持 GET（列出工具）和 POST（JSON-RPC）。

### `Server.Export(method, path string, handler app.HandlerFunc, opts ...ToolOption) error`

直接在服务器实例上导出工具的便捷方法。

## 配置选项

| 选项 | 说明 |
|------|------|
| `WithEndpoint(path string)` | 自定义 MCP 端点路径（默认：`/mcp`） |
| `WithName(name string)` | 自定义工具名称（默认：根据 method + path 自动生成） |
| `WithDescription(desc string)` | 设置工具描述 |
| `WithInputSchema(schema map[string]interface{})` | 覆盖默认输入模式 |

## 项目结构

```
mcp/
├── mcp.go          # 入口：New(), Export()
├── server.go       # MCP 服务器：dispatch、initialize、tools/list、tools/call
├── registry.go     # 线程安全的工具注册表
├── executor.go     # 工具执行引擎
├── naming.go       # 默认工具名称生成
├── options.go      # 配置选项
├── *_test.go       # 单元测试和集成测试
└── example/
    └── basic/
        └── main.go # 最小可运行示例
```
