package ElevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
)

func existsRequestsAbove(elev Elevator) bool {
	for floor := elev.Floor + 1; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if elev.Requests[floor][button] {
				return true
			}
		}
	}
	return false
}

func existsRequestsBelow(elev Elevator) bool {
	for floor := 0; floor < elev.Floor; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if elev.Requests[floor][button] {
				return true
			}
		}
	}
	return false
}

func chooseDirection(elev Elevator) elevio.MotorDirection {
	switch elev.Direction {

	case elevio.MD_Stop:
		fallthrough

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
	}
	return elevio.MD_Stop
}

func shouldStop(elev Elevator) bool {
	var floor int = elev.Floor

	switch elev.Direction {
	case elevio.MD_Up:
		return elev.Requests[floor][elevio.BT_HallUp] ||
			elev.Requests[floor][elevio.BT_Cab] ||
			!existsRequestsAbove(elev)

	case elevio.MD_Down:
		return elev.Requests[floor][elevio.BT_HallDown] ||
			elev.Requests[floor][elevio.BT_Cab] ||
			!existsRequestsBelow(elev)

	}
	return true
}
