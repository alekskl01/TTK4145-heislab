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

func cyclicCounter(requests [config.N_FLOORS][config.N_BUTTONS]RequestState, order elevio.ButtonEvent, otherState []RequestState) {
	floor := order.Floor
	button_type := order.Button

	myState := requests[floor][button_type]

	switch myState {
	case NoRequest:
		if otherCountersAhead(PendingRequest, otherState) {
			requests[floor][button_type] = PendingRequest
		}

	case PendingRequest:
		if checkEqualityForArray(myState, otherState) {
			requests[floor][button_type] = ActiveRequest
			// TODO: Turn on button light somewhere

		} else if otherCountersAhead(ActiveRequest, otherState) {
			requests[floor][button_type] = ActiveRequest
			// TODO: Turn on button light somewhere
		}

	case ActiveRequest:
		if otherCountersAhead(DeleteRequest, otherState) {
			requests[floor][button_type] = DeleteRequest
		}

	case DeleteRequest:
		if checkEqualityForArray(myState, otherState) {
			requests[floor][button_type] = NoRequest
			// TODO: Turn off button light somewhere

		} else if otherCountersAhead(ActiveRequest, otherState) {
			requests[floor][button_type] = NoRequest
			// TODO: Turn off button light somewhere
		}

	}
}
