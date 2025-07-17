package httponly

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/ppipada/go-mcp-expt/jsonrpc/humaadapter"
	jsonrpcReqResp "github.com/ppipada/go-mcp-expt/jsonrpc/reqresp"
)

const (
	JSONRPCEndpoint = "/jsonrpc"
)

func Register(api huma.API,
	methodMap map[string]jsonrpcReqResp.IMethodHandler,
	notificationMap map[string]jsonrpcReqResp.INotificationHandler,
) {
	// Get default operation.
	op := humaadapter.GetDefaultOperation()
	op.Path = JSONRPCEndpoint
	// Register the methods.
	humaadapter.Register(api, op, methodMap, notificationMap, nil, nil)
}
