package halldelegator

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/request"
	"fmt"
	"math"
)

func log(text string) {
	fmt.Println("Hall Delegator: " + text)
}

func getCostBetweenFloors(current_floor int, floor int) (int, elevio.MotorDirection) {
	var difference = current_floor - floor
	// Unless obstructed we define 0 cost to move to current floor.
	if difference == 0 {
		return 0, elevio.MD_Stop
	}
	var direction = elevio.MD_Up
	if difference < 0 {
		difference = difference * -1
		direction = elevio.MD_Down
	}
	return difference, direction
}

// Gives an approximate "cost" of taking a hall order for an elevator.
// Being perfectly accurate and efficient is not important, while consistency is.
func GetCostOfHallOrder(hall_floor int, button_type elevio.ButtonType, floor int, direction elevio.MotorDirection,
	is_obstructed bool, requests [config.N_FLOORS][config.N_BUTTONS]request.RequestState) int {
	if button_type == elevio.BT_Cab {
		// Assume some kind of unexpected bug, ensure cost is always highest
		log("Tried to calculate cost of a cab order.")
		return config.HIGH_COST
	}

	// Costs nothing to take a hall order we already have taken.
	if request.IsActive(requests[hall_floor][button_type]) {
		return config.LOW_COST
	}

	var cost int = 0
	if is_obstructed {
		cost = cost + config.MAJOR_COST
	}
	hall_distance, hall_dir := getCostBetweenFloors(floor, hall_floor)
	// Unless obstructed we define 0 cost to move to current floor.
	if hall_distance == 0 {
		return cost
	}

	cost = cost + hall_distance	
	for request_floor := 0; request_floor < config.N_FLOORS; request_floor++ {
		floor_distance, floor_dir := getCostBetweenFloors(floor, request_floor)
		// Any active request in the direction we need to go means less cost while
		// requests in the opposite direction mean additional cost.
		// This does not take into account the difference between a HallUp and HallDown request
		if request.IsActive(requests[request_floor][elevio.BT_Cab]) ||
			request.IsActive(requests[request_floor][elevio.BT_HallDown]) ||
			request.IsActive(requests[request_floor][elevio.BT_HallUp]) {
			if floor_dir == hall_dir {
				cost = cost - floor_distance
			} else {
				cost = cost + floor_distance
			}
		}
	}

	// Add or subtract 2 since that is the number of floor changes needed to 
	// be at the same floor with the opposite direction.
	if (direction == hall_dir) {
		cost = cost - 2
	} else {
		cost = cost + 2
	}

	return cost
}
