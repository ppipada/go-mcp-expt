package humaadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	jsonrpcReqResp "github.com/ppipada/go-mcp-expt/jsonrpc/reqresp"
)

// GetDefaultOperation gets the conventional values for jsonrpc as a single operation.
func GetDefaultOperation() huma.Operation {
	return huma.Operation{
		Method:        http.MethodPost,
		Path:          "/jsonrpc",
		DefaultStatus: 200,

		Tags:        []string{"JSONRPC"},
		Summary:     "JSONRPC endpoint",
		Description: "Serve all jsonrpc methods",
		OperationID: "jsonrpc",
	}
}

// Response object. It implements the huma StatusError interface.
// IF the JSONRPC handler is invoked, it should never throw an error, but should return a error response object.
// JSONRPC requires a error case to be covered via the specifications error response object.
func GetErrorHandler(
	methodMap map[string]jsonrpcReqResp.IMethodHandler,
	notificationMap map[string]jsonrpcReqResp.INotificationHandler,
) func(status int, message string, errs ...error) huma.StatusError {
	return func(gotStatus int, gotMessage string, errs ...error) huma.StatusError {
		var foundJSONRPCError *jsonrpcReqResp.JSONRPCError
		message := gotMessage
		details := make([]string, 0)
		// Add the Message and HTTP status to details and set status sent back as 200.
		details = append(details, "Message:"+gotMessage, "HTTP Status:"+strconv.Itoa(gotStatus))
		status := 200

		code := jsonrpcReqResp.InternalError
		if gotStatus >= 400 && gotStatus < 500 {
			code = jsonrpcReqResp.InvalidRequestError
			message = jsonrpcReqResp.GetDefaultErrorMessage(jsonrpcReqResp.InvalidRequestError)
		}

		for _, err := range errs {
			var jsonRPCError jsonrpcReqResp.JSONRPCError
			if converted, ok := err.(huma.ErrorDetailer); ok {
				d := converted.ErrorDetail()
				// See if this is parse error.
				if strings.Contains(d.Message, "unmarshal") ||
					strings.Contains(d.Message, "invalid character") ||
					strings.Contains(d.Message, "unexpected end") {
					code = jsonrpcReqResp.ParseError
					message = jsonrpcReqResp.GetDefaultErrorMessage(jsonrpcReqResp.ParseError)
				}
			} else if errors.As(err, &jsonRPCError) {
				// Check if the error is of type JSONRPCError.
				foundJSONRPCError = &jsonRPCError
			}
			details = append(details, err.Error())
		}

		// If a JSONRPCError was found, update the message and append JSON-encoded details.
		if foundJSONRPCError != nil {
			message = foundJSONRPCError.Message
			code = foundJSONRPCError.Code

			// JSON encode the Data field of the found JSONRPCError.
			if jsonData, err := json.Marshal(foundJSONRPCError.Data); err == nil {
				details = append(details, string(jsonData))
			}
		}

		// Check for method not found.
		if gotMessage == "validation failed" {
			// Assume that the method name is in one of the error messages
			// Look for "method:<methodName>".
			var methodName string
			for _, errMsg := range details {
				idx := strings.Index(errMsg, "method:")
				if idx != -1 {
					// Extract method name up to the next space or bracket or end of string.
					rest := errMsg[idx+len("method:"):]
					endIdx := strings.IndexFunc(rest, func(r rune) bool {
						return r == ' ' || r == ']' || r == ')'
					})
					if endIdx == -1 {
						methodName = rest
					} else {
						methodName = rest[:endIdx]
					}
					break
				}
			}
			// Check if method exists in methodMap or notificationMap.
			if methodName != "" {
				if _, exists := methodMap[methodName]; !exists {
					if _, exists := notificationMap[methodName]; !exists {
						// Method not found.
						// You need to define this constant.
						code = jsonrpcReqResp.MethodNotFoundError
						message = fmt.Sprintf("Method '%s' not found", methodName)
					}
				}
			}
		}

		return &ResponseStatusError{
			status: status,
			Response: jsonrpcReqResp.Response[any]{
				JSONRPC: jsonrpcReqResp.JSONRPCVersion,
				ID:      nil,
				Error: &jsonrpcReqResp.JSONRPCError{
					Code:    code,
					Message: message,
					Data:    details,
				},
			},
		}
	}
}

// Register a new JSONRPC operation.
// The `methodMap` maps from method name to request handlers. Request clients expect a response object
// The `notificationMap` maps from method name to notification handlers. Notification clients do not expect a response
//
// These maps can be instantiated as
//
//	methodMap := map[string]jsonrpc.IMethodHandler{
//		"add": &jsonrpc.MethodHandler[AddParams, int]{Endpoint: AddEndpoint},
//	}
//
//	notificationMap := map[string]jsonrpc.INotificationHandler{
//		"log": &jsonrpc.NotificationHandler[LogParams]{Endpoint: LogEndpoint},
//	}
func Register(
	api huma.API,
	op huma.Operation,
	methodMap map[string]jsonrpcReqResp.IMethodHandler,
	notificationMap map[string]jsonrpcReqResp.INotificationHandler,
	responseMap map[string]jsonrpcReqResp.IResponseHandler,
	responseHandlerMapper func(context.Context, jsonrpcReqResp.Response[json.RawMessage]) (string, error),
) {
	AddSchemasToAPI(api, methodMap, notificationMap)
	huma.NewError = GetErrorHandler(methodMap, notificationMap)
	brh := jsonrpcReqResp.NewBatchRequestHandler(jsonrpcReqResp.WithMethodMap(methodMap),
		jsonrpcReqResp.WithNotificationMap(notificationMap),
		jsonrpcReqResp.WithResponseMap(responseMap, responseHandlerMapper),
	)

	huma.Register(api, op, brh.Handle)
}
