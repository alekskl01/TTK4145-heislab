// This file manages and keeps track of local requests, with some use of network information.
// See the request package / file for more details on how individual requests are structured.
package statemanager

import (
	"Elevator/config"
	"Elevator/elevatorstate"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
)

func existsRequestsOnFloor(elev *elevatorstate.Elevator) bool {
	for button := 0; button < config.N_BUTTONS; button++ {
		if request.IsActive(elev.Requests[elev.Floor][button]) && SelfAssignedRequests[elev.Floor][button] {
			return true
		}
	}
	return false
}

func existsRequestsAbove(elev *elevatorstate.Elevator) bool {
	for floor := elev.Floor + 1; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if request.IsActive(elev.Requests[floor][button]) && SelfAssignedRequests[floor][button] {
				return true
			}
		}
	}
	return false
}

func existsRequestsBelow(elev *elevatorstate.Elevator) bool {
	for floor := 0; floor < elev.Floor; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if request.IsActive(elev.Requests[floor][button]) && SelfAssignedRequests[floor][button] {
				return true
			}
		}
	}
	return false
}

func existsRequests(elev *elevatorstate.Elevator) bool {
	return (existsRequestsAbove(elev) || existsRequestsBelow(elev) || existsRequestsOnFloor(elev))
}

func chooseDirection(elev *elevatorstate.Elevator) elevio.MotorDirection {
	switch elev.Direction {

	case elevio.MD_Stop:
		fallthrough

	case elevio.MD_Up:
		{
			if existsRequestsAbove(elev) {
				return elevio.MD_Up
			} else if existsRequestsBelow(elev) {
				return elevio.MD_Down
			} else {
				return elevio.MD_Stop
			}
		}

	case elevio.MD_Down:
		{
			if existsRequestsBelow(elev) {
				return elevio.MD_Down
			} else if existsRequestsAbove(elev) {
				return elevio.MD_Up
			} else {
				return elevio.MD_Stop
			}
		}
	}
	return elevio.MD_Stop
}

func shouldStop(elev *elevatorstate.Elevator) bool {
	var floor int = elev.Floor

	switch elev.Direction {
	case elevio.MD_Up:
		return request.IsActive(elev.Requests[floor][elevio.BT_HallUp]) ||
			request.IsActive(elev.Requests[floor][elevio.BT_Cab]) ||
			!existsRequestsAbove(elev)

	case elevio.MD_Down:
		return request.IsActive(elev.Requests[floor][elevio.BT_HallDown]) ||
			request.IsActive(elev.Requests[floor][elevio.BT_Cab]) ||
			!existsRequestsBelow(elev)

	}
	return true
}

// Worth noting that this function takes into account the requests of other nodes,
// in order to properly trigger deletion for all elevators.
func clearRequestAtFloor(elev *elevatorstate.Elevator) {
	var servicedHallRequest elevio.ButtonType

	if request.IsActive(elev.Requests[elev.Floor][elevio.BT_HallDown]) &&
		request.IsActive(elev.Requests[elev.Floor][elevio.BT_HallUp]) {

		switch elev.Direction {
		case elevio.MD_Stop:
			fallthrough

		case elevio.MD_Up:
			servicedHallRequest = elevio.BT_HallUp

		case elevio.MD_Down:
			servicedHallRequest = elevio.BT_HallDown
		}
	} else if request.IsActive(elev.Requests[elev.Floor][elevio.BT_HallDown]) {
		servicedHallRequest = elevio.BT_HallDown
	} else if request.IsActive(elev.Requests[elev.Floor][elevio.BT_HallUp]) {
		servicedHallRequest = elevio.BT_HallUp
	}

	var otherStates, _ = network.GetRequestStatesAtIndex(elev.Floor, elevio.BT_Cab)
	if request.OrderStatesEqualTo(request.ActiveRequest, elev.Requests[elev.Floor][elevio.BT_Cab], otherStates) {
		elev.Requests[elev.Floor][elevio.BT_Cab] = request.DeleteRequest
	}
	elevio.SetButtonLamp(elevio.ButtonType(elevio.BT_Cab), elev.Floor, false)

	otherStates, _ = network.GetRequestStatesAtIndex(elev.Floor, servicedHallRequest)
	if request.OrderStatesEqualTo(request.ActiveRequest, elev.Requests[elev.Floor][servicedHallRequest], otherStates) {
		elev.Requests[elev.Floor][servicedHallRequest] = request.DeleteRequest
	}
	elevio.SetButtonLamp(elevio.ButtonType(servicedHallRequest), elev.Floor, false)
}
