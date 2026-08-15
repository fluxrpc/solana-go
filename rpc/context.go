package rpc

// Context is the RPC response context, describing the state the response
// was evaluated at.
type Context struct {
	Slot       uint64  `json:"slot"`
	ApiVersion *string `json:"apiVersion,omitempty"`
}

// RPCContext wraps the response context, embedded by "withContext" results.
type RPCContext struct {
	Context Context `json:"context,omitempty"`
}

// M is a generic JSON object, used for free-form RPC parameters and values.
type M map[string]any
