package requests

import (
	"Elevator/elevatorfsm"
	"Elevator/elevio"
)

const NUM_BUTTONS int = 3

func requestAbove(elev elevatorfsm.Elevator) bool {
	for floor := elev.Floor + 1; floor < elevio.NUM_FLOORS; floor++ {
		for button := 0; button < NUM_BUTTONS; button++ {
			if elev.Requests[floor][button] == 1 {
				return true
			}
		}
	}
	return false
}

func requestBelow(elev elevatorfsm.Elevator) bool {
	for floor := 0; floor < elev.Floor; floor++ {
		for button := 0; button < NUM_BUTTONS; button++ {
			if elev.Requests[floor][button] == 1 {
				return true
			}
		}
	}
	return false
}

func chooseDirection(elev elevatorfsm.Elevator) elevio.MotorDirection {
	switch elev.Direction {
	case elevio.MD_Up:
		{
			if requestAbove(elev) {
				return elevio.MD_Up
			} else if requestBelow(elev) {
				return elevio.MD_Down
			} else {
				return elevio.MD_Stop
			}
		}
	case elevio.MD_Down:
		{
			if requestBelow(elev) {
				return elevio.MD_Down
			} else if requestAbove(elev) {
				return elevio.MD_Up
			} else {
				return elevio.MD_Stop
			}
		}
	default:
		return elevio.MD_Stop
	}
}

func shouldStop(elev elevatorfsm.Elevator) bool {
	var floor int = elev.Floor

	switch elev.Direction {
	case elevio.MD_Up:
		return elev.Requests[floor][elevio.BT_HallUp] == 1 ||
			elev.Requests[floor][elevio.BT_Cab] == 1 ||
			!requestAbove(elev)

	case elevio.MD_Down:
		return elev.Requests[floor][elevio.BT_HallDown] == 1 ||
			elev.Requests[floor][elevio.BT_Cab] == 1 ||
			!requestBelow(elev)

	default:
		return true
	}
}

func clearRequestsAtFloor(elev elevatorfsm.Elevator) {
	for button := 0; button < NUM_BUTTONS; button++ {
		elev.Requests[elev.Floor][button] = 0
	}
}

func clearAllRequests(elev elevatorfsm.Elevator) {
	for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
		for button := 0; button < NUM_BUTTONS; button++ {
			elev.Requests[elev.Floor][button] = 0
		}
	}
}
