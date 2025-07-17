package spec

import (
	jsonrpcReqResp "github.com/ppipada/go-mcp-expt/jsonrpc/reqresp"
)

// A request from the client to the server, to ask for completion options.
type CompleteRequest jsonrpcReqResp.Request[CompleteRequestParams]

type CompleteRequestParams struct {
	// This property is reserved by the protocol to allow clients and servers
	// to attach additional metadata to their requests.
	Meta     map[string]any                `json:"_meta,omitempty"`
	_        struct{}                      `json:"-"               additionalProperties:"true"`
	Argument CompleteRequestParamsArgument `json:"argument"`

	// Ref corresponds to the JSON schema field "ref".
	Ref CompletionReference `json:"ref"`
}

// The argument's information.
type CompleteRequestParamsArgument struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Identifies a completion.
type CompletionReference struct {
	Type Ref `json:"type"`
	// This should be present for prompt.
	Name *string `json:"name"`
	// This should be present for resource.
	URI *string `json:"uri"`
}

// The server's response to a completion/complete request.
type (
	CompleteResponse jsonrpcReqResp.Response[*CompleteResult]
	CompleteResult   struct {
		// This result property is reserved by the protocol to allow clients and servers
		// to attach additional metadata to their responses.
		Meta map[string]any `json:"_meta,omitempty"`

		// Completion corresponds to the JSON schema field "completion".
		Completion CompleteResultCompletion `json:"completion"`
	}
)

type CompleteResultCompletion struct {
	// Indicates whether there are additional completion options beyond those provided
	// in the current response, even if the exact total is unknown.
	HasMore *bool `json:"hasMore,omitempty"`

	// The total number of completion options available. This can exceed the number of
	// values actually sent in the response.
	Total *int `json:"total,omitempty"`

	// An array of completion values. Must not exceed 100 items.
	Values []string `json:"values"`
}
