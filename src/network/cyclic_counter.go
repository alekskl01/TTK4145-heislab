package network

import (
	"Elevator/elevatorFSM"
	"Elevator/request"
)

func checkEqualityForArray(myState RequestRequest, otherState []request.RequestState) bool {
	for _, RequestState := range otherState {
		if RequestState != myState {
			return false
		}
	}
	return true
}

func otherCountersAhead(myNextState request.RequestState, otherState []request.RequestState) bool {
	for _, state := range otherState {
		if state == myNextState {
			return true
		}
	}
	return false
}

func orderStatesEqualTo(checkState request.RequestState, myState request.RequestState, otherState []request.RequestState) {
	if myState == checkState {
		if checkEqualityForArray(checkState, otherState) {
			return true
		}
	}
	return false
}

func cyclicCounter(elev *elevatorFSM.Elevator, order ButtonEvent, otherState []request.RequestState) {
	floor := order.Floor
	button_type := order.Button

	myState := elev.Requests[floor][button_type]

	switch request.myOrderState {
	case request.NoRequest:
		if otherCountersAhead(myState, otherState, request.PendingRequest) {
			elev.Requests[floor][button_type] = request.PendingRequest
		}

	case request.PendingRequest:
		if checkEqualityForArray(myState, otherState) {
			elev.Requests[floor][button_type] = request.ActiveRequest
			// TODO: Turn on button light somewhere

		} else if otherCountersAhead(myState, otherState, request.ActiveRequest) {
			elev.Requests[floor][button_type] = request.ActiveRequest
			// TODO: Turn on button light somewhere
		}

	case request.ActiveRequest:
		if otherCountersAhead(myState, otherState, request.DeleteRequest) {
			elev.Requests[floor][button_type] = request.DeleteRequest
		}

	case request.DeleteRequest:
		if checkEqualityForArray(myState, otherState) {
			elev.Requests[floor][button_type] = request.NoRequest
			// TODO: Turn off button light somewhere

		} else if otherCountersAhead(myState, otherState, request.ActiveRequest) {
			elev.Requests[floor][button_type] = request.NoRequest
			// TODO: Turn off button light somewhere
		}

	}
}
