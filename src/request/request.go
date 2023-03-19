package request

type RequestState int

const (
	NoRequest RequestState = iota
	PendingRequest
	ActiveRequest
	DeleteRequest
)

func Ac(state RequestState) bool {
	return state == ActiveRequest
}
