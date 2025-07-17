package humaadapter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	jsonrpcReqResp "github.com/ppipada/go-mcp-expt/jsonrpc/reqresp"
)

type (
	RequestAny      = jsonrpcReqResp.Request[any]
	NotificationAny = jsonrpcReqResp.Notification[any]
)

func IntStringSchema() *huma.Schema {
	return &huma.Schema{
		OneOf: []*huma.Schema{
			{Type: huma.TypeInteger},
			{Type: huma.TypeString},
		},
	}
}

func getTypeSchema(
	api huma.API,
	methodName string,
	mtype reflect.Type,
	suffix string,
) *huma.Schema {
	hint := methodName + suffix
	inputSchema := api.OpenAPI().Components.Schemas.Schema(mtype, true, hint)
	if inputSchema.Ref != "" {
		inputSubSchema := api.OpenAPI().Components.Schemas.SchemaFromRef(inputSchema.Ref)
		inputSubSchema.Title = inputSchema.Ref[strings.LastIndex(inputSchema.Ref, "/")+1:]
	} else if hint != "" {
		// Base types
		// E.g: For string param huma name will be String and hint will be the above.
		// For Array.
		humaName := huma.DefaultSchemaNamer(mtype, hint)
		titlecaseHint := strings.ToUpper(string(hint[0])) + hint[1:]
		if titlecaseHint != humaName {
			inputSchema.Title = titlecaseHint + " - " + humaName
		} else {
			inputSchema.Title = titlecaseHint
		}
	}

	return inputSchema
}

func isNillableType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	default:
		return false
	}
}

// Function to dynamically create the Request type with Params of type iType.
func getRequestType(iType reflect.Type, isNotification bool) reflect.Type {
	// Get the reflect.Type of Request[any].
	requestAnyType := reflect.TypeOf(RequestAny{})
	if isNotification {
		requestAnyType = reflect.TypeOf(NotificationAny{})
	}

	// Get the number of fields in the Request struct.
	numFields := requestAnyType.NumField()

	// Create a slice to hold the StructField definitions.
	fields := make([]reflect.StructField, numFields)

	// Iterate over each field in the Request struct.
	for i := range numFields {
		// Get the field.
		field := requestAnyType.Field(i)

		// If the field is 'Params', replace its type with iType.
		if field.Name == "Params" {
			field.Type = iType
			jsonTag := field.Tag.Get("json")

			// If iType is pointer type, add omitempty to json tag.
			if isNillableType(iType) {
				if !strings.Contains(jsonTag, "omitempty") {
					jsonTag += ",omitempty"
				}
				// Update the field's tag with the modified JSON and required tags.
				field.Tag = reflect.StructTag(
					fmt.Sprintf(`json:%q`, jsonTag),
				)
			} else {
				// If iType is not a pointer, add required:true to required tag
				// Update the field's tag with the modified JSON and required tags.
				field.Tag = reflect.StructTag(
					fmt.Sprintf(`json:%q required:"true"`, jsonTag),
				)
			}
		}

		// Add the field to the fields slice.
		fields[i] = field
	}
	// Create a new struct type with the updated fields.
	reqType := reflect.StructOf(fields)
	return reqType
}

func getRequestSchema(
	api huma.API,
	methodName string,
	paramType reflect.Type,
	isNotification bool,
) *huma.Schema {
	newReqType := getRequestType(paramType, isNotification)
	reqSchema := getTypeSchema(api, methodName, newReqType, "Request")
	if reqSchema.Properties == nil {
		reqSchema.Properties = make(map[string]*huma.Schema)
	}
	if reqSchema.Ref != "" {
		reqSubSchema := api.OpenAPI().Components.Schemas.SchemaFromRef(reqSchema.Ref)
		// Set method name as a constant in the schema.
		reqSubSchema.Properties["method"] = &huma.Schema{
			Type: "string",
			Enum: []any{methodName},
		}
		if !isNotification {
			reqSubSchema.Properties["id"] = IntStringSchema()
			reqSubSchema.Required = append(reqSubSchema.Required, "id")
		}
	}
	return reqSchema
}

func getResponseSchema(
	api huma.API,
	methodName string,
	paramType reflect.Type,
) *huma.Schema {
	// Get the error type used in your application.
	errorType := reflect.TypeOf(jsonrpcReqResp.JSONRPCError{})

	// Create dynamic types for success and error responses.
	successResponseType := getSuccessResponseType(paramType)
	errorResponseType := getErrorResponseType(errorType)

	// Generate schemas for these dynamic types.
	successSchema := getTypeSchema(
		api,
		methodName,
		successResponseType,
		"SuccessResponse",
	)
	if successSchema.Ref != "" {
		reqSubSchema := api.OpenAPI().Components.Schemas.SchemaFromRef(successSchema.Ref)
		reqSubSchema.Properties["id"] = IntStringSchema()
		reqSubSchema.Required = append(reqSubSchema.Required, "id")
	}
	errorSchema := getTypeSchema(api, methodName, errorResponseType, "ErrorResponse")
	if errorSchema.Ref != "" {
		reqSubSchema := api.OpenAPI().Components.Schemas.SchemaFromRef(errorSchema.Ref)
		reqSubSchema.Properties["id"] = IntStringSchema()
		reqSubSchema.Required = append(reqSubSchema.Required, "id")
	}

	// Build the response schema with OneOf combining the two schemas.
	responseSchema := &huma.Schema{
		Title: strings.ToUpper(string(methodName[0])) + methodName[1:] + "Response",
		OneOf: []*huma.Schema{
			successSchema,
			errorSchema,
		},
	}

	return responseSchema
}

// Function to create the success response type dynamically.
func getSuccessResponseType(resultType reflect.Type) reflect.Type {
	fields := []reflect.StructField{
		{
			Name: "Jsonrpc",
			Type: reflect.TypeOf(""),
			Tag:  `json:"jsonrpc"`,
		},
		{
			Name: "Id",
			Type: reflect.TypeOf((*jsonrpcReqResp.IntString)(nil)).Elem(),
			Tag:  `json:"id"`,
		},
	}
	var resultField reflect.StructField
	resultField.Name = "Result"
	resultField.Type = resultType

	if isNillableType(resultType) {
		// If resultType is a pointer, add omitempty to json tag.
		resultField.Tag = reflect.StructTag(`json:"result,omitempty"`)
	} else {
		resultField.Tag = reflect.StructTag(`json:"result" required:"true"`)
	}

	fields = append(fields, resultField)

	return reflect.StructOf(fields)
}

// Function to create the error response type dynamically.
func getErrorResponseType(errorType reflect.Type) reflect.Type {
	fields := []reflect.StructField{
		{
			Name: "Jsonrpc",
			Type: reflect.TypeOf(""),
			Tag:  `json:"jsonrpc"`,
		},
		{
			Name: "Id",
			Type: reflect.TypeOf((*jsonrpcReqResp.IntString)(nil)).Elem(),
			Tag:  `json:"id"`,
		},
		{
			Name: "Error",
			Type: errorType,
			Tag:  `json:"error"`,
		},
	}
	return reflect.StructOf(fields)
}

func AddSchemasToAPI(
	api huma.API,
	methodMap map[string]jsonrpcReqResp.IMethodHandler,
	notificationMap map[string]jsonrpcReqResp.INotificationHandler,
) {
	// Prepare slices to hold per-method request and response schemas.
	reqSchemas := make([]*huma.Schema, 0, len(methodMap)+len(notificationMap))
	resSchemas := make([]*huma.Schema, 0, len(methodMap))

	// Process method handlers.
	for methodName, handler := range methodMap {
		inputType, outputType := handler.GetTypes()

		reqSchema := getRequestSchema(api, methodName, inputType, false)
		reqSchemas = append(reqSchemas, reqSchema)

		respSchema := getResponseSchema(api, methodName, outputType)
		resSchemas = append(resSchemas, respSchema)
	}

	// Process notification handlers.
	for methodName, handler := range notificationMap {
		inputType := handler.GetTypes()
		reqSchema := getRequestSchema(api, methodName, inputType, true)
		reqSchemas = append(reqSchemas, reqSchema)
	}

	// Get base Request[json.RawMessage] and Response[json.RawMessage] schemas.
	reqType := reflect.TypeOf((*jsonrpcReqResp.Request[json.RawMessage])(nil)).Elem()
	baseReqSchema := api.OpenAPI().Components.Schemas.Schema(reqType, false, "")
	baseReqSchema.OneOf = reqSchemas
	// Delete properties.
	baseReqSchema.Properties = make(map[string]*huma.Schema)
	baseReqSchema.Required = []string{}
	baseReqSchema.AdditionalProperties = true
	baseReqSchema.Type = ""

	respType := reflect.TypeOf((*jsonrpcReqResp.Response[json.RawMessage])(nil)).Elem()
	baseRespSchema := api.OpenAPI().Components.Schemas.Schema(respType, false, "")
	baseRespSchema.OneOf = resSchemas
	// Delete properties.
	baseRespSchema.Properties = make(map[string]*huma.Schema)
	baseRespSchema.Required = []string{}
	baseRespSchema.AdditionalProperties = true
	baseRespSchema.Type = ""

	elementType := reflect.TypeOf((*jsonrpcReqResp.UnionRequest)(nil)).Elem()
	elementSchema := api.OpenAPI().Components.Schemas.Schema(elementType, true, "")
	if elementSchema.Ref != "" {
		elementSubSchema := api.OpenAPI().Components.Schemas.SchemaFromRef(elementSchema.Ref)
		elementSubSchema.Properties["id"] = IntStringSchema()
	}
	batchItemType := reflect.TypeOf((*jsonrpcReqResp.BatchItem[jsonrpcReqResp.UnionRequest])(nil)).
		Elem()
	basebatchItemSchema := api.OpenAPI().Components.Schemas.Schema(batchItemType, false, "")
	*basebatchItemSchema = huma.Schema{
		OneOf: []*huma.Schema{
			elementSchema, {
				Type:     huma.TypeArray,
				Items:    elementSchema,
				MinItems: intPtr(1),
			},
		},
	}

	batchItemResponseType := reflect.TypeOf((*jsonrpcReqResp.BatchItem[jsonrpcReqResp.Response[json.RawMessage]])(nil)).
		Elem()
	basebatchItemResponseSchema := api.OpenAPI().Components.Schemas.Schema(
		batchItemResponseType,
		false,
		"",
	)
	*basebatchItemResponseSchema = huma.Schema{
		OneOf: []*huma.Schema{
			baseRespSchema, {
				Type:     huma.TypeArray,
				Items:    baseRespSchema,
				MinItems: intPtr(1),
			},
		},
	}
}

func intPtr(i int) *int {
	return &i
}
