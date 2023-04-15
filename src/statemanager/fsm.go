// This file contains the main state machine that governs the behaviour of the local elevator
package statemanager

import (
	"Elevator/config"
	es "Elevator/elevatorstate" // Acronym alias used to shorten repeated code
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
	"fmt"
)

// Contains a continously updated overview of which requests are cheapest
// for the local node to take,
var SelfAssignedRequests [config.N_FLOORS][config.N_BUTTONS]bool

func log(text string) {
	fmt.Println("FSM: " + text)
}

func InitSelfAssignedRequests() {
	// Cab orders are always cheapest for us to take
	for floor := 0; floor < config.N_FLOORS; floor++ {
		SelfAssignedRequests[floor][config.N_BUTTONS-1] = true
	}
}

func RunStateMachine(elevator *es.Elevator, buttonPressEventCh <-chan elevio.ButtonEvent, floorArrivalEventCh <-chan int,
	obstructionEventCh <-chan bool, requestsUpdatedCh <-chan [config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	for {
		select {
		case order := <-buttonPressEventCh:
			floor := order.Floor
			buttonType := order.Button
			var otherStates, _ = network.GetRequestStatesAtIndex(floor, buttonType)

			// Prevents taking hall orders if we are not connected to the network, part of ensuring proper resynchronization
			if len(otherStates) == 0 && buttonType != elevio.BT_Cab {
				break
			}

			switch elevator.State {
			case es.DoorOpen:
				if elevator.Floor == floor {
					// If the elevator is already at the floor where the order was placed, we only add the order if it is a hall order,
					// otherwise we just open the door again.
					if ((buttonType == elevio.BT_HallUp && elevator.Direction == elevio.MD_Down) ||
						(buttonType == elevio.BT_HallDown && elevator.Direction == elevio.MD_Up)) &&
						request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {

						elevator.Requests[floor][buttonType] = request.PendingRequest
					}
					openDoor(elevator)
					elevator.State = es.DoorOpen
				} else if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
					elevator.Requests[floor][buttonType] = request.PendingRequest
				}

			case es.Moving:
				if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
					elevator.Requests[floor][buttonType] = request.PendingRequest
				}

			case es.MotorStop:
				if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
					elevator.Requests[floor][buttonType] = request.PendingRequest
				}

			case es.Idle:
				if elevator.Floor == floor {
					// See comment above
					if ((buttonType == elevio.BT_HallUp && elevator.Direction == elevio.MD_Down) ||
						(buttonType == elevio.BT_HallDown && elevator.Direction == elevio.MD_Up)) &&
						request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {

						elevator.Requests[floor][buttonType] = request.PendingRequest
					}
					openDoor(elevator)
					elevator.State = es.DoorOpen
				} else {
					if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
						elevator.Requests[floor][buttonType] = request.PendingRequest
					}
				}
			}
			es.SetButtonLights(elevator)

		case newFloor := <-floorArrivalEventCh:
			elevator.Floor = newFloor
			elevio.SetFloorIndicator(newFloor)

			if elevator.State == es.Moving || elevator.State == es.MotorStop {

				if elevator.State == es.MotorStop {
					elevator.State = es.Moving
				}
				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevator.MotorStopTimer.Stop()

					clearRequestAtFloor(elevator)
					es.SetButtonLights(elevator)

					openDoor(elevator)
					elevator.State = es.DoorOpen
				} else {
					elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}
			}

		case obstruction := <-obstructionEventCh:
			elevator.Obstruction = obstruction

			if !elevator.DoorTimer.HasTimeRemaining() && !elevator.Obstruction {
				onDoorTimeout(elevator)
			}

		case <-elevator.DoorTimer.Timer.C:
			if !elevator.Obstruction {
				onDoorTimeout(elevator)
			}

		case <-elevator.MotorStopTimer.Timer.C:
			log("Triggered motor stop timer ")
			switch elevator.State {
			case es.Moving:
				elevator.State = es.MotorStop
				// Adds a pending request if there is no other pending request to ensure it arrives at a valid floor
				if !existsRequests(elevator) {
					var otherStates, _ = network.GetRequestStatesAtIndex(elevator.Floor+int(elevator.Direction), elevio.BT_Cab)
					if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[elevator.Floor+int(elevator.Direction)][elevio.BT_Cab], otherStates) {
						elevator.Requests[elevator.Floor+int(elevator.Direction)][elevio.BT_Cab] = request.PendingRequest
					}
				}
				elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)

			case es.MotorStop:
				elevator.Direction = chooseDirection(elevator)
				elevio.SetMotorDirection(elevator.Direction)
				elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
			}

		case updatedRequests := <-requestsUpdatedCh:
			elevator.Requests = updatedRequests
			es.SetButtonLights(elevator)

			// Start servicing newly confirmed requests (When in idle state).
			if elevator.State == es.Idle {
				if existsRequestsOnFloor(elevator) {
					clearRequestAtFloor(elevator)
					openDoor(elevator)
					elevator.State = es.DoorOpen
				} else if existsRequestsBelow(elevator) || existsRequestsAbove(elevator) {
					elevator.Direction = chooseDirection(elevator)
					elevio.SetMotorDirection(elevator.Direction)
					elevator.State = es.Moving
					elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}
			}
		}
	}
}

func onDoorTimeout(elevator *es.Elevator) {
	floor := elevator.Floor

	// Checks if elevator has a reason to travel in direction decided when floor
	// was reached if unserviced hall orders exist on the current floor, and services the hall order if not.

	if elevator.State == es.DoorOpen && !elevator.Obstruction {
		if request.IsActive(elevator.Requests[floor][elevio.BT_HallDown]) {
			if !existsRequestsAbove(elevator) {

				var otherStates, _ = network.GetRequestStatesAtIndex(floor, elevio.BT_HallDown)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallDown], otherStates) {
					elevator.Requests[floor][elevio.BT_HallDown] = request.DeleteRequest
				}
				es.SetButtonLights(elevator)
				openDoor(elevator)
				elevator.State = es.DoorOpen
				return

			} else {
				var otherStates, _ = network.GetRequestStatesAtIndex(floor, elevio.BT_HallUp)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallUp], otherStates) {
					elevator.Requests[floor][elevio.BT_HallUp] = request.DeleteRequest
				}
			}
		} else if request.IsActive(elevator.Requests[elevator.Floor][elevio.BT_HallUp]) {
			if !existsRequestsBelow(elevator) {

				var otherStates, _ = network.GetRequestStatesAtIndex(floor, elevio.BT_HallUp)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallUp], otherStates) {
					elevator.Requests[floor][elevio.BT_HallUp] = request.DeleteRequest
				}
				es.SetButtonLights(elevator)
				openDoor(elevator)
				elevator.State = es.DoorOpen
				return

			} else {
				var otherStates, _ = network.GetRequestStatesAtIndex(floor, elevio.BT_HallDown)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallDown], otherStates) {
					elevator.Requests[floor][elevio.BT_HallDown] = request.DeleteRequest
				}
			}
		}

		elevio.SetDoorOpenLamp(false)
		elevator.Direction = chooseDirection(elevator)
		elevio.SetMotorDirection(elevator.Direction)

		if elevator.Direction == elevio.MD_Stop {
			elevator.State = es.Idle
		} else {
			elevator.State = es.Moving
			elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
		}
	}
}

func openDoor(elevator *es.Elevator) {
	elevio.SetDoorOpenLamp(true)
	elevator.DoorTimer.Reset(config.DOOR_OPEN_DURATION)
}
