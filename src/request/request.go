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
