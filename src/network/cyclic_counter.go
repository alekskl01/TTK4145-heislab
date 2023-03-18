package network

import (
	"Elevator/ElevatorFSM"
)

type RequestState int

const (
	NoRequest   RequestState = iota
	PendingRequest                
	ActiveRequest
	DeleteRequest              
)

func checkEqualityForArray(myState RequestRequest, otherState []RequestState) bool {
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

func orderStatesEqualTo(checkState RequestState, myState RequestState, otherState []RequestState) {
	if myState == checkState {
		if checkEqualityForArray(checkState, otherState) {
			return true
		}
	}
	return false
}


func cyclicCounter(elev &ElevatorFSM.Elevator, order ButtonEvent, otherState []RequestState) {
	floor := order.Floor
	button_type := order.Button

	myState := elev.Requests[floor][button_type]

	switch myOrderState {
	case NoRequest:
		if otherCountersAhead(myState, otherState, PendingRequest) {
			elev.Requests[floor][button_type] = PendingRequest
		}
	
	case PendingRequest:
		if checkEqualityForArray(myState, otherState) {
			elev.Requests[floor][button_type] = ActiveRequest
			// TODO: Turn on button light somewhere

		} else if otherCountersAhead(myState, otherState, ActiveRequest) {
			elev.Requests[floor][button_type] = ActiveRequest
			// TODO: Turn on button light somewhere
		}
	
	case ActiveRequest:
		if otherCountersAhead(myState, otherState, DeleteRequest) {
			elev.Requests[floor][button_type] = DeleteRequest
		}

	case DeleteRequest:
		if checkEqualityForArray(myState, otherState) {
			elev.Requests[floor][button_type] = NoRequest
			// TODO: Turn off button light somewhere

		} else if otherCountersAhead(myState, otherState, ActiveRequest) {
			elev.Requests[floor][button_type] = NoRequest
			// TODO: Turn off button light somewhere
		}
	

	}
}
