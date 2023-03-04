package ElevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
)

func ExistsRequestAbove(elev Elevator) bool {
	for floor := elev.floor + 1; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if elev.requests[floor][button] == true {
				return true
			}
		}
	}
	return false
}

func ExistsRequestBelow(elev Elevator) bool {
	for floor := 0; floor < elev.floor; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if elev.requests[floor][button] == true {
				return true
			}
		}
	}
	return false
}

func ExistRequestsAtFloor(elev Elevator)

func chooseDirection(elev Elevator) elevio.MotorDirection {
	switch elev.direction {
	case elevio.MD_Up:
		{
			if ExistsRequestAbove(elev) {
				return elevio.MD_Up
			} else if ExistsRequestBelow(elev) {
				return elevio.MD_Down
			} else {
				return elevio.MD_Stop
			}
		}
	case elevio.MD_Down:
		{
			if ExistsRequestBelow(elev) {
				return elevio.MD_Down
			} else if ExistsRequestAbove(elev) {
				return elevio.MD_Up
			} else {
				return elevio.MD_Stop
			}
		}
	default:
		return elevio.MD_Stop
	}
}

func ShouldStop(elev Elevator) bool {
	var floor int = elev.floor

	switch elev.direction {
	case elevio.MD_Up:
		return elev.requests[floor][elevio.BT_HallUp] == true ||
			elev.requests[floor][elevio.BT_Cab] == true ||
			!ExistsRequestAbove(elev)

	case elevio.MD_Down:
		return elev.requests[floor][elevio.BT_HallDown] == true ||
			elev.requests[floor][elevio.BT_Cab] == true ||
			!ExistsRequestBelow(elev)

	default:
		return true
	}
}
