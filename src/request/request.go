// This package and file concerns the handling of specific requests that represent
// active orders and allow for them to be synchronized and fulfilled in a predictable manner.
// The term 'request' refers to a potential hall or cab order, therefore the terms are used interchangeably.
package request

type RequestState int

const (
	// No orders of that type given.
	NoRequest RequestState = iota
	// Specific order given but not synchronized.
	PendingRequest
	// Order synchronized, such that actions (lights, fulfillment) can be performed.
	ActiveRequest
	// Order fulfillment pending synchronization.
	DeleteRequest
)

// Shorthand which also helps to ensure that actions are only performed after synchronization.
func IsActive(state RequestState) bool {
	return state == ActiveRequest
}

// Used to determine whether the light for a request should be on
func ShouldActivateButtonLight(state RequestState) bool {
	return (state == ActiveRequest || state == DeleteRequest)
}
