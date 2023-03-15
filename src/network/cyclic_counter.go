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

func cyclicCounter(elev ElevatorFSM.Elevator, myOrderState RequestState, otherOrderState []RequestState) {
	switch myOrderState {
	case NoRequest:
		for _, RequestState := range otherOrderState {
			if RequestState == RecievedRequest {
				elev.Requests[][]
			}
		}
	
	case RecievedRequest:
		break
	
	case ActiveRequest:
		break

	case DeleteRequest:
		break

	}
}