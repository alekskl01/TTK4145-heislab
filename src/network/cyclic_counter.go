package network

import (
	"Elevator/ElevatorFSM"
)

type RequestState int

const (
	NoRequest   RequestState = iota
	RecievedRequest                
	ActiveRequest
	DeleteRequest              
)

func checkEqualityForArray(myState RecievedRequest, otherState []RequestState) bool {
	for _, RequestState := range otherState {
		if RequestState != myState {
			return false
		}
	}
	return true
}

func otherCountersAhead(myState RequestState, otherState []RequestState, nextState RequestState) bool {
	for _, state := range otherOrderState {
		if state == nextState {
			return true
		}
	}
	return false
}


func cyclicCounter(elev &ElevatorFSM.Elevator, order ButtonEvent, otherOrderState []RequestState) {
	floor := order.Floor
	button_type := order.Button

	myState := elev.Requests[floor][button_type]

	switch myOrderState {
	case NoRequest:
		if otherCountersAhead(myState, otherOrderState, RecievedRequest) {
			elev.Requests[floor][button_type] = RecievedRequest
		}
	
	case RecievedRequest:
		if checkEqualityForArray(myState, otherOrderState) {
			elev.Requests[floor][button_type] = ActiveRequest
			// TODO: Turn on button light somewhere

		} else if otherCountersAhead(myState, otherOrderState, ActiveRequest) {
			elev.Requests[floor][button_type] = ActiveRequest
			// TODO: Turn on button light somewhere
		}
	
	case ActiveRequest:
		if otherCountersAhead(myState, otherOrderState, DeleteRequest) {
			elev.Requests[floor][button_type] = DeleteRequest
		}

	case DeleteRequest:
		if checkEqualityForArray(myState, otherOrderState) {
			elev.Requests[floor][button_type] = NoRequest
			// TODO: Turn off button light somewhere

		} else if otherCountersAhead(myState, otherOrderState, ActiveRequest) {
			elev.Requests[floor][button_type] = NoRequest
			// TODO: Turn off button light somewhere
		}
	

	}
}