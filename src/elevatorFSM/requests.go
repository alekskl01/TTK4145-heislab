package elevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
)

func existsRequestsAbove(elev *Elevator) bool {
	for floor := elev.Floor + 1; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if request.IsActive(elev.Requests[floor][button]) {
				return true
			}
		}
	}
	return false
}

func existsRequestsBelow(elev *Elevator) bool {
	for floor := 0; floor < elev.Floor; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if request.IsActive(elev.Requests[floor][button]) {
				return true
			}
		}
	}
	return false
}

func chooseDirection(elev *Elevator) elevio.MotorDirection {
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

func shouldStop(elev *Elevator) bool {
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

func clearRequestAtFloor(elev *Elevator) {

	nextDirection := chooseDirection(elev)
	var servicedHallRequest elevio.ButtonType

	switch nextDirection {
	case elevio.MD_Stop:
		fallthrough

	case elevio.MD_Up:
		servicedHallRequest = elevio.BT_HallUp

	case elevio.MD_Down:
		servicedHallRequest = elevio.BT_HallDown
	}

	var otherStates = network.GetRequestStatesAtIndex(elev.Floor, elevio.BT_Cab)
	//TODO: improve code quality
	if request.OrderStatesEqualTo(request.ActiveRequest, elev.Requests[elev.Floor][elevio.BT_Cab], otherStates) {
		elev.Requests[elev.Floor][elevio.BT_Cab] = request.DeleteRequest
	}
	elevio.SetButtonLamp(elevio.ButtonType(elevio.BT_Cab), elev.Floor, false)
	otherStates = network.GetRequestStatesAtIndex(elev.Floor, servicedHallRequest)
	if request.OrderStatesEqualTo(request.ActiveRequest, elev.Requests[elev.Floor][servicedHallRequest], otherStates) {
		elev.Requests[elev.Floor][servicedHallRequest] = request.DeleteRequest
	}
	elevio.SetButtonLamp(elevio.ButtonType(servicedHallRequest), elev.Floor, false)
}
