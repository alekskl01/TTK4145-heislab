package requests

import "../driver"

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
	for floor:=0; floor =< elev.floor ; floor++ {
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
	case MD_Up:
		if requestAbove(elev) {
			return elevio.MD_Up
		}
		else if requestBelow(elev) {
			return elevio.MD_Down
		}
		else {
			elevio.MD_Stop
		}
	}
	case MD_Down:
		if requestBelow(elev) {
			return elevio.
		}


}