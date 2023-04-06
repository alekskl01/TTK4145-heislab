package synchronizer

import (
	"Elevator/config"
	"Elevator/elevatorFSM"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
	"time"
)

func LocalRequestSynchronization(elev *elevatorFSM.Elevator, requestsUpdate chan<- [config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
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
		time.Sleep(250 * time.Millisecond)
	}
}

// func GlobalRequestSynchronization() {
// 	for {
// 		for _, n := range network.ConnectedNodes {
// 			for floor := 0; floor < config.N_FLOORS; floor++ {
// 				var otherStates = network.GetRequestStatesAtIndex(floor, elevio.BT_Cab)
// 				var existingSyncState = network.GlobalElevatorStates[n]
// 				existingSyncState.LocalRequests[floor][elevio.BT_Cab] = request.CyclicCounter(network.GlobalElevatorStates[n].LocalRequests, floor, elevio.BT_Cab, otherStates)
// 				network.GlobalElevatorStates[n] = existingSyncState
// 			}
// 		}
// 		time.Sleep(250 * time.Millisecond)
// 	}
// }
