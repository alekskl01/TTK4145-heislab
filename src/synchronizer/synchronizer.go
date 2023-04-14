package synchronizer

import (
	"Elevator/config"
	"Elevator/elevatorstate"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
	"time"
)

func LocalRequestSynchronization(elev *elevatorstate.Elevator, requestsUpdate chan<- [config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	for {
		for floor := 0; floor < config.N_FLOORS; floor++ {
			for button := 0; button < config.N_BUTTONS; button++ {
				var otherStates = network.GetRequestStatesAtIndex(floor, elevio.ButtonType(button))
				var newState = request.CyclicCounter(elev.Requests, floor, elevio.ButtonType(button), otherStates)
				if elev.Requests[floor][button] != newState {
					var newRequests = elev.Requests
					newRequests[floor][button] = newState
					requestsUpdate <- newRequests
				}
			}
		}
		time.Sleep(config.UPDATE_DELAY)
	}
}

func UpdateCheapestRequests(floor *int, direction *elevio.MotorDirection,
	is_obstructed *bool, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	for {
		for hall_floor := 0; hall_floor < config.N_FLOORS; hall_floor++ {
			for button := 0; button < (config.N_BUTTONS - 1); button++ {
				elevatorstate.CheapestRequests[hall_floor][button] = network.IsHallOrderCheapest(hall_floor, elevio.ButtonType(button), floor, direction, is_obstructed, requests)
			}
		}
		time.Sleep(config.UPDATE_DELAY)
	}
}
