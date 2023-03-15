package requestSync

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
		break
	
	case RecievedRequest:
		break
	
	case ActiveRequest:
		break

	case DeleteRequest:
		break
		
	}
}