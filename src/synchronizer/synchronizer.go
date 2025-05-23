// Continously updates the state of all local requests
// and which requests are cheapest for the local node to take.
package synchronizer

import (
	"Elevator/config"
	"Elevator/elevatorstate"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
	"Elevator/statemanager"
	"time"
)

// Continously iterates through indivdual request counters based on all network request states.
// Note that counter incrementation is paused when a node is in the process of resynchronization.
func LocalRequestSynchronization(elev *elevatorstate.Elevator, requestsUpdate chan<- [config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	for {
		for floor := 0; floor < config.N_FLOORS; floor++ {
			for button := 0; button < config.N_BUTTONS; button++ {
				var otherStates, anyNotSynchronized = network.GetRequestStatesAtIndex(floor, elevio.ButtonType(button))
				if !anyNotSynchronized {
					var newState = request.CyclicCounter(elev.Requests, floor, elevio.ButtonType(button), otherStates)
					if elev.Requests[floor][button] != newState {
						var newRequests = elev.Requests
						newRequests[floor][button] = newState
						requestsUpdate <- newRequests
					}
				}
			}
		}
		time.Sleep(config.UPDATE_DELAY)
	}
}

func UpdateSelfAssignedRequests(floor *int, direction *elevio.MotorDirection, fsm_state *elevatorstate.ElevatorState,
	is_obstructed *bool, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	for {
		for hall_floor := 0; hall_floor < config.N_FLOORS; hall_floor++ {
			for button := 0; button < (config.N_BUTTONS - 1); button++ {
				statemanager.SelfAssignedRequests[hall_floor][button] = network.IsHallOrderCheapest(hall_floor, elevio.ButtonType(button), floor, direction, fsm_state, is_obstructed, requests)
			}
		}
		time.Sleep(config.UPDATE_DELAY)
	}
}
