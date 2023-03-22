package synchronizer

import (
	"Elevator/config"
	"Elevator/elevatorFSM"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
)

func LocalRequestSynchronization(elev *elevatorFSM.Elevator) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			var otherStates = network.GetRequestStatesAtIndex(floor, elevio.ButtonType(button))
			elev.Requests[floor][button] = request.CyclicCounter(elev.Requests, floor, elevio.ButtonType(button), otherStates)
		}
	}
}

func GlobalRequestSynchronization() {
	for _, n := range network.ConnectedNodes {
		for floor := 0; floor < config.N_FLOORS; floor++ {
			var otherStates = network.GetRequestStatesAtIndex(floor, elevio.BT_Cab)
			var existingSyncState = network.NetworkRequests[n]
			existingSyncState.Requests[floor][elevio.BT_Cab] = request.CyclicCounter(network.NetworkRequests[n].Requests, floor, elevio.BT_Cab, otherStates)			
			network.NetworkRequests[n] = existingSyncState
		}
	}
}
