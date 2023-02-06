package requests

import "Elevator/elevio"

func requestAbove(elev Elevator) bool {
	for floor := elev.floor + 1; floor < NUM_FLOORS; floor++ {
		for button := 0; button < NUM_BUTTONS; button++ {
			if elev.requests[floor][button] {
				return true
			}
		}
	}
	return false
}

func requestBelow(elev Elevator) bool {
	for floor:=0; floor < elev.floor ; floor++ {
		for button := 0; button < NUM_BUTTONS; button++ {
			if elev.requests[floor][button] {
				return true
			}
		}
	}
	return false
}

func chooseDirection(elev Elevator) elevio.MotorDirection {
	switch elev.direction {
	case MD_Up: {
		if requestAbove(elev) {
			return elevio.MD_Up
		} else if requestBelow(elev) {
			return elevio.MD_Down
		} else {
			elevio.MD_Stop
		}
	}
	case MD_Down: {
		if requestBelow(elev) {
			return elevio.MD_Down
		} else if requestAbove(elev) {
			return elevio.MD_Up
		} else {
			return MD_Stop
		}
	}
	default:
		return MD_Stop
	}
}

func shouldStop(elev Elevator) bool {
	var floor int = elev.floor

	switch elev.direction {
	case MD_Up:
		return elev.requests[floor][elevio.BT_HallUp] ||
			elev.requests[floor][elevio.BT_Cab] ||
			!requestAbove(elev) 

	case MD_Down:
		return elev.requests[floor][elevio.BT_HallDown] ||
			elev.requests[floor][elevio.BT_Cab] ||
			!requestBelow(elev)

	default:
		return true
	}
}

func clearRequestsAtFloor(elev) {
	for button := 0; button < NUM_BUTTONS {
		elev.requests[elev.foor][button] = 0;
	} 
}

func clearAllRequests(elev) {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for button := 0; button < NUM_BUTTONS; button++ {
			elev.requests[elev.foor][button] = 0;
		}
	} 
}

