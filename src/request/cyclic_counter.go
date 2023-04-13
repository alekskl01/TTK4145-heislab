package request

import (
	"Elevator/config"
	"Elevator/elevio"
)

func checkEqualityForArray(myState RequestState, otherState []RequestState) bool {
	for _, RequestState := range otherState {
		if RequestState != myState {
			return false
		}
	}
	return true
}

func otherCountersAhead(myNextState RequestState, otherState []RequestState) bool {
	for _, state := range otherState {
		if state == myNextState {
			return true
		}
	}
	return false
}

func OrderStatesEqualTo(checkState RequestState, myState RequestState, otherState []RequestState) bool {
	if myState == checkState {
		if checkEqualityForArray(checkState, otherState) {
			return true
		}
	}
	return false
}

func CyclicCounter(requests [config.N_FLOORS][config.N_BUTTONS]RequestState, floor int, button_type elevio.ButtonType, otherState []RequestState) RequestState {

	myState := requests[floor][button_type]
	
	switch myState {
	case NoRequest:
		if otherCountersAhead(PendingRequest, otherState) {
			return PendingRequest
		}

	case PendingRequest:
		if checkEqualityForArray(myState, otherState) {
			return ActiveRequest
			// TODO: Turn on button light somewhere

		} else if otherCountersAhead(ActiveRequest, otherState) {
			return ActiveRequest
			// TODO: Turn on button light somewhere
		}

	case ActiveRequest:
		if otherCountersAhead(DeleteRequest, otherState) {
			return DeleteRequest
		}

	case DeleteRequest:
		if checkEqualityForArray(myState, otherState) {
			return NoRequest
			// TODO: Turn off button light somewhere

		} else if otherCountersAhead(NoRequest, otherState) {
			return NoRequest
			// TODO: Turn off button light somewhere
		}

	}
	return myState
}
