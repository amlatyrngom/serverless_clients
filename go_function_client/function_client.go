package function_client

import "database/sql"

// Client to interact with function manager.
type FunctionClient struct {
	db *sql.DB
}

// Wrapper to call functions.
type Function struct {
	FunctionID int64
}

// Info returned by invocation. The result is asynchronously placed in the channel.
type InvocationInfo struct {
	Err          error
	InvocationID int64
	ResCh        chan InvocationResult
}

// Result of an invocation.
type InvocationResult struct {
	InvocationID int64
	Res          []byte
	Err          error
}

// Return new client
func NewClient(db_host string, db_port string) (*FunctionClient, error) {
	return nil, nil
}

// Get a function.
func (function_client *FunctionClient) Get(function_id int64) (*Function, error) {
	function := &Function{
		FunctionID: function_id,
	}
	return function, nil
}

// Perform invocation.
func (function *Function) Invoke(params []byte) *InvocationInfo {
	return nil
}

// Interrupt invocation.
func (function *Function) Interrupt(invocation_id int64) error {
	return nil
}
