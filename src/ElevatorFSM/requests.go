package ElevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
)

func existsRequestsAbove(elev Elevator) bool {
	for floor := elev.floor + 1; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if elev.requests[floor][button] == true {
				return true
			}
		}
	}
	return false
}

func existsRequestsBelow(elev Elevator) bool {
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
			if existsRequestsAbove(elev) {
				return elevio.MD_Up
			} else if existsRequestsBelow(elev) {
				return elevio.MD_Down
			} else {
				return elevio.MD_Stop
			}
		}
	case elevio.MD_Down:
		{
			if existsRequestsBelow(elev) {
				return elevio.MD_Down
			} else if existsRequestsAbove(elev) {
				return elevio.MD_Up
			} else {
				return elevio.MD_Stop
			}
		}
	default:
		return elevio.MD_Stop
	}
}

func shouldStop(elev Elevator) bool {
	var floor int = elev.floor

	switch elev.direction {
	case elevio.MD_Up:
		return elev.requests[floor][elevio.BT_HallUp] == true ||
			elev.requests[floor][elevio.BT_Cab] == true ||
			!existsRequestsAbove(elev)

	case elevio.MD_Down:
		return elev.requests[floor][elevio.BT_HallDown] == true ||
			elev.requests[floor][elevio.BT_Cab] == true ||
			!existsRequestsBelow(elev)

	default:
		return true
	}
}
