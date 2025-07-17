package helpers_test

import (
	"context"

	jsonrpcReqResp "github.com/ppipada/go-mcp-expt/jsonrpc/reqresp"
)

// AddParams defines the parameters for the "add" method.
type AddParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddResult struct {
	Sum int `json:"sum"`
}

type NotifyParams struct {
	Message string `json:"message"`
}

// ConcatParams defines the parameters for the "concat" method.
type ConcatParams struct {
	S1 string `json:"s1"`
	S2 string `json:"s2"`
}

// PingParams defines the parameters for the "ping" notification.
type PingParams struct {
	Message string `json:"message"`
}

// AddEndpoint is the handler for the "add" method.
func AddEndpoint(ctx context.Context, params AddParams) (AddResult, error) {
	res := params.A + params.B
	return AddResult{Sum: res}, nil
}

// ConcatEndpoint is the handler for the "concat" method.
func ConcatEndpoint(ctx context.Context, params ConcatParams) (string, error) {
	return params.S1 + params.S2, nil
}

// PingEndpoint is the handler for the "ping" notification.
func PingEndpoint(ctx context.Context, params PingParams) error {
	return nil
}

func NotifyEndpoint(ctx context.Context, params NotifyParams) error {
	return nil
}

func GetMethodHandlers() map[string]jsonrpcReqResp.IMethodHandler {
	methodMap := map[string]jsonrpcReqResp.IMethodHandler{
		"add": &jsonrpcReqResp.MethodHandler[AddParams, AddResult]{Endpoint: AddEndpoint},
		"addpositional": &jsonrpcReqResp.MethodHandler[[]int, AddResult]{
			Endpoint: func(ctx context.Context, params []int) (AddResult, error) {
				res := 0
				for _, v := range params {
					res += v
				}
				return AddResult{Sum: res}, nil
			},
		},
		"concat": &jsonrpcReqResp.MethodHandler[ConcatParams, string]{Endpoint: ConcatEndpoint},
		"concatOptionalIn": &jsonrpcReqResp.MethodHandler[*ConcatParams, string]{
			Endpoint: func(ctx context.Context, params *ConcatParams) (string, error) {
				if params != nil {
					return params.S1 + params.S2, nil
				}
				return "", nil
			},
		},
		"concatOptionalInOut": &jsonrpcReqResp.MethodHandler[*ConcatParams, *string]{
			Endpoint: func(ctx context.Context, params *ConcatParams) (*string, error) {
				r := ""
				if params != nil {
					r = params.S1 + params.S2
					return &r, nil
				}
				return &r, nil
			},
		},
		"echo": &jsonrpcReqResp.MethodHandler[any, any]{
			Endpoint: func(ctx context.Context, x any) (any, error) {
				return x, nil
			},
		},
		"echooptional": &jsonrpcReqResp.MethodHandler[*string, *string]{
			Endpoint: func(ctx context.Context, e *string) (*string, error) {
				return e, nil
			},
		},
	}

	return methodMap
}

func GetNotificationHandlers() map[string]jsonrpcReqResp.INotificationHandler {
	notificationMap := map[string]jsonrpcReqResp.INotificationHandler{
		"ping": &jsonrpcReqResp.NotificationHandler[PingParams]{Endpoint: PingEndpoint},
		"notify": &jsonrpcReqResp.NotificationHandler[NotifyParams]{
			Endpoint: NotifyEndpoint,
		},
	}

	return notificationMap
}
