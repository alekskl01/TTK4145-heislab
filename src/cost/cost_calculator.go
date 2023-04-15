// The job of the this package is calculating a comparable 'cost'
// of fulfilling hall orders for specific nodes, so that the ideal node
// for fulfulling an order can be determined.
package cost

import (
	"Elevator/config"
	"Elevator/elevatorstate"
	"Elevator/elevio"
	"Elevator/request"
	"fmt"
)

func log(text string) {
	fmt.Println("Cost calculator: " + text)
}

// Gives the direction and distance needed to get from one floor to another.
func getNeededFloorTraversal(current_floor int, target_floor int) (int, elevio.MotorDirection) {
	var difference = current_floor - target_floor
	// Unless obstructed we define 0 cost to move to current floor.
	if difference == 0 {
		return 0, elevio.MD_Stop
	}
	direction := elevio.MotorDirection(elevio.MD_Down)
	if difference < 0 {
		difference = difference * -1
		direction = elevio.MD_Up
	}
	return difference, direction
}

// Gives an approximate "cost" of taking a hall order for an elevator.
// Being perfectly accurate and efficient is not important, while consistency is.
func GetCostOfHallOrder(hallFloor int, button_type elevio.ButtonType, floor int, direction elevio.MotorDirection,
	fsm_state elevatorstate.ElevatorState, is_obstructed bool, requests [config.N_FLOORS][config.N_BUTTONS]request.RequestState) int {
	if button_type == elevio.BT_Cab {
		// Assume some kind of unexpected bug, ensure cost is always highest
		log("Tried to calculate cost of a cab order.")
		return config.HIGH_COST
	}

	// Costs nothing to take a hall order we already have taken.
	if request.IsActive(requests[hallFloor][button_type]) {
		return config.LOW_COST
	}

	var cost = 0
	if is_obstructed {
		cost = cost + config.MAJOR_COST
	}

	// Stopped motor can't fulfill orders.
	if fsm_state == elevatorstate.MotorStop {
		return config.HIGH_COST
	}

	hallDistance, hallDir := getNeededFloorTraversal(floor, hallFloor)
	// Unless obstructed we define 0 cost to move to current floor.
	if hallDistance == 0 {
		return cost
	}

	if fsm_state == elevatorstate.DoorOpen {
		// Smallest possible increment to give minor benefit to closed door elevators
		cost = cost + 1
	}

	cost = cost + hallDistance
	for requestFloor := 0; requestFloor < config.N_FLOORS; requestFloor++ {
		floorDistance, floor_dir := getNeededFloorTraversal(floor, requestFloor)
		// Any active request in the direction we need to go means less cost while
		// requests in the opposite direction mean additional cost.
		// This does not take into account the difference between a HallUp and HallDown request
		if request.IsActive(requests[requestFloor][elevio.BT_Cab]) ||
			request.IsActive(requests[requestFloor][elevio.BT_HallDown]) ||
			request.IsActive(requests[requestFloor][elevio.BT_HallUp]) {
			if floor_dir == hallDir {
				cost = cost - floorDistance
			} else {
				cost = cost + floorDistance
			}
		}
	}
	
	// Add or subtract 2 since that is the number of floor changes needed to change directions
	// Easier to move in direction we are already moving in.
	if direction != elevio.MD_Stop {
		if direction == hallDir {
			cost = cost - 2
		} else {
			cost = cost + 2
		}
	}

	return cost
}
