# 测试指南

## 运行全部测试

```bash
cd mcp
go test ./...
```

## 测试文件说明

| 文件 | 覆盖范围 |
|------|----------|
| `mcp_test.go` | `New()`、`Export()` — 服务器创建、参数校验、选项覆盖 |
| `registry_test.go` | `Registry` — 并发注册、重复检测、列表查询 |
| `naming_test.go` | `defaultToolName()` — 根据 HTTP 方法和路径生成工具名 |
| `executor_test.go` | `Executor` — 工具执行、工具不存在错误、上下文复用 |
| `server_test.go` | `Server` — 请求分发、initialize、tools/list、tools/call、错误处理 |
| `integration_test.go` | 端到端 HTTP 测试 — 从工具注册到调用的完整生命周期 |

## 核心测试用例

### 服务器测试（`server_test.go`）

- **`TestServerHandlerList`** — GET `/mcp` 返回工具列表
- **`TestServerHandlerCall`** — POST `/mcp` 通过 `tools/call` 调用处理器，返回 SSE 响应
- **`TestServerHandlerInitialize`** — `initialize` 请求返回协议版本、能力和服务器信息
- **`TestServerHandlerInitializedNotification`** — `notifications/initialized` 通知被正确接受
- **`TestServerHandlerMethodNotAllowed`** — 不支持的 HTTP 方法返回 405
- **`TestServerHandlerInvalidJSON`** — 格式错误的请求体返回 JSON-RPC 解析错误（-32700）

### 集成测试（`integration_test.go`）

- **`TestIntegrationEndToEnd`** — 启动真实的 Hertz 服务器，注册工具后：
  1. GET `/mcp` 验证工具列表
  2. POST `/mcp` 通过 `tools/call` 调用处理器
  3. 验证 SSE 响应格式和工具输出

## MCP 协议合规验证

所有测试均验证以下协议要求：

- 响应使用 SSE 格式（`text/event-stream`）
- 成功响应包含 `resultType: "complete"`
- 错误响应遵循 JSON-RPC 错误码（-32700、-32601、-32602、-32000）
- `initialize` 返回 `protocolVersion`、`capabilities` 和 `serverInfo`

## curl 测试

### 1. 启动示例服务器

```bash
cd mcp/example/basic
go run main.go
```

### 2. 获取工具列表

```bash
curl http://localhost:8080/mcp
```

### 3. Initialize 握手

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-11-25",
      "capabilities": {},
      "clientInfo": {"name": "curl-test", "version": "1.0.0"}
    }
  }'
```

### 4. 调用工具

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "get_v1_user_test",
      "arguments": {
        "path": {"id": "42"},
        "query": {"verbose": "true"}
      }
    }
  }'
```

## MCP Inspector 测试

### 1. 安装并启动 Inspector

```bash
npx @modelcontextprotocol/inspector
```

### 2. 启动 MCP 服务器

```bash
cd mcp/example/basic
go run main.go
```

### 3. 连接 Inspector

- 打开浏览器访问 `http://localhost:5173`（Inspector 默认端口）
- 选择 **HTTP** 传输模式
- 输入服务器地址：`http://localhost:8080/mcp`
- 点击 **Connect**
- Inspector 会自动发送 `initialize` 握手，成功后即可浏览工具列表和调用工具

## AI IDE MCP 测试（如 Trae / Cursor / Claude Desktop）

### 1. 配置 MCP 服务器

在 AI IDE 的 MCP 配置文件中添加：

```json
{
  "mcpServers": {
    "hertz-mcp": {
      "url": "http://localhost:8080/mcp",
      "transport": "http"
    }
  }
}
```

> **注意**：部分客户端使用 SSE 传输，需确认服务器支持。当前实现通过 POST `/mcp` 返回 SSE 格式响应，兼容 HTTP 模式的 MCP 客户端。

### 2. 启动 MCP 服务器

```bash
cd mcp/example/basic
go run main.go
```

### 3. 验证连接

- 重启 AI IDE 或重新加载 MCP 配置
- 在 IDE 的 MCP 工具面板中应能看到 `get_v1_user_test` 工具
- 通过自然语言指令让 AI 调用工具，例如："调用 get_v1_user_test 获取 id 为 42 的用户"

### 4. 调试技巧

- 如果工具未显示，检查服务器日志确认是否收到 `initialize` 请求
- 使用 curl 先验证服务器是否正常响应
- 确认 MCP 配置文件路径正确（Trae: `.trae/mcp.json`，Cursor: `.cursor/mcp.json`）
