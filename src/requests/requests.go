package requests

import (
	"Elevator/elevator"
	"Elevator/elevio"
)

func ExistsRequestAbove(elev elevator.Elevator) bool {
	for floor := elev.Floor + 1; floor < elevio.NUM_FLOORS; floor++ {
		for button := 0; button < elevio.NUM_BUTTONS; button++ {
			if elev.Requests[floor][button] == 1 {
				return true
			}
		}
	}
	return false
}

func ExistsRequestBelow(elev elevator.Elevator) bool {
	for floor := 0; floor < elev.Floor; floor++ {
		for button := 0; button < elevio.NUM_BUTTONS; button++ {
			if elev.Requests[floor][button] == 1 {
				return true
			}
		}
	}
	return false
}

func ExistRequestsAtFloor(elev elevator.Elevator)

func ChooseDirection(elev elevator.Elevator) elevio.MotorDirection {
	switch elev.Direction {
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

func ShouldStop(elev elevator.Elevator) bool {
	var floor int = elev.Floor

	switch elev.Direction {
	case elevio.MD_Up:
		return elev.Requests[floor][elevio.BT_HallUp] == 1 ||
			elev.Requests[floor][elevio.BT_Cab] == 1 ||
			!ExistsRequestAbove(elev)

	case elevio.MD_Down:
		return elev.Requests[floor][elevio.BT_HallDown] == 1 ||
			elev.Requests[floor][elevio.BT_Cab] == 1 ||
			!ExistsRequestBelow(elev)

	default:
		return true
	}
}
