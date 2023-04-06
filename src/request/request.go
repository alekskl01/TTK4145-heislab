package request

type RequestState int

const (
	NoRequest RequestState = iota
	PendingRequest
	ActiveRequest
	DeleteRequest
)

func IsActive(state RequestState) bool {
	return state == ActiveRequest
}

// Used to determine whether the light for a request should be on
func ShouldActivateButtonLight(state RequestState) bool {
	return state == ActiveRequest || state == DeleteRequest
}
