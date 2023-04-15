// This file contains the methods for managing the cyclic counter
// representing the state of a particular request
package request

import (
	"Elevator/config"
	"Elevator/elevio"
	"fmt"
)

func log(text string) {
	fmt.Println("Cyclic counter: " + text)
}

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

// Describes sequential iteration of request states to keep orders synchronized and allow them to be fulfilled.
// The lifecycle of a request is incremented as it is synchronized and fulfilled.
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

		} else if otherCountersAhead(ActiveRequest, otherState) {
			return ActiveRequest
		}

	case ActiveRequest:
		if otherCountersAhead(DeleteRequest, otherState) {
			return DeleteRequest
		}

	case DeleteRequest:
		if checkEqualityForArray(myState, otherState) {
			return NoRequest

		} else if otherCountersAhead(NoRequest, otherState) {
			return NoRequest
		}

	}
	return myState
}
