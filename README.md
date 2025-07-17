# Go MCP Experiments

> ⚠️ **Experimental project - development is currently on hold.**  
> APIs are unstable and subject to change at any time.

A playground for building a layered [Model-Context-Protocol (MCP)](https://modelcontextprotocol.io/introduction) SDK in Go.
The repo is split into three mostly independent layers:

1. Dependency-free JSON-RPC core
2. Pluggable transports: Stdio, HTTP/SSE, Plain HTTP. A [Huma](https://github.com/danielgtaylor/huma) example server is also present.
3. Go-idiomatic snapshot of the MCP schema

## 1. JSON-RPC Layer

| Feature          | Notes                                                                                          | Status |
| ---------------- | ---------------------------------------------------------------------------------------------- | ------ |
| Meta-handler     | Inspects raw JSON and dispatches `method`, `notification` or `response` with batching support. | Stable |
| Handler registry | Simple `func(ctx, params) (result, error)` signature                                           | Stable |
| Huma adapter     | Exposes any JSON-RPC server as an OpenAPI endpoint via Huma                                    | Stable |

## 2. Pluggable Transports

| Transport    | Highlights                                                                            | Status                        |
| ------------ | ------------------------------------------------------------------------------------- | ----------------------------- |
| `Stdio`      | Dependency-free `net.Conn` wrapper + server that mirrors `http.HandlerFunc` semantics | Functional. Needs MCP updates |
| `HTTP + SSE` | Tunnels JSON-RPC over Server-Sent Events                                              | Functional. Needs MCP updates |
| `Plain HTTP` | Minimal example of non-MCP JSON-RPC over HTTP                                         | Example only                  |

## 3. MCP SDK

- Based on an older snapshot of the official MCP schema.
- Converted to idiomatic Go to cope with heavy use of unions/overloads in the base schema.
- JSON-schema conversion is complete; mapping that spec onto the transports is **incomplete**.

## Dev

- Prerequisites

- `Go >= 1.24`
- `golangci-lint >= 2.x`

- Setup

  - `git clone https://github.com/ppipada/go-mcp-expt`
  - `cd go-mcp-expt`
  - `go mod download`

- Common tasks
  - Lint : `golangci-lint run ./... -v`
  - Tests: `go test ./...`
  - Run a local Huma server exposing JSON-RPC as OpenAPI: `./run_jsonrpc.sh`

## Roadmap / TODO

- [ ] Update transports to the latest MCP spec
- [ ] Add MCP method handlers.

## License

This project is licensed under the **MIT License**. See [LICENSE](./LICENSE) for details.
