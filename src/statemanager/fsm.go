// This file contains the main state machine that governs the behaviour of the local elevator
package statemanager

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
	es "Elevator/elevatorstate" // Acronym alias used to shorten repeated code
	"fmt"
	"strconv"
)

var CheapestRequests [config.N_FLOORS][config.N_BUTTONS]bool

func log(text string) {
	fmt.Println("FSM: " + text)
}

func InitCheapestRequests() {
	// Cab orders are always cheapest for us to take
	for floor := 0; floor < config.N_FLOORS; floor++ {
		CheapestRequests[floor][config.N_BUTTONS-1] = true
	}
}

func RunStateMachine(elevator *es.Elevator, buttonPressEventCh <-chan elevio.ButtonEvent, floorArrivalEventCh <-chan int,
	obstructionEventCh <-chan bool, stopButtonEventCh <-chan bool, elevatorUnavailableCh chan<- bool, requestsUpdatedCh <-chan [config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	for {
		select {
		case order := <-buttonPressEventCh:
			floor := order.Floor
			buttonType := order.Button
			var otherStates, _ = network.GetRequestStatesAtIndex(floor, buttonType)

			// Prevents taking hall orders if we are not connected to the network
			if len(otherStates) == 0 && buttonType != elevio.BT_Cab {
				break
			}

			switch elevator.State {
			case es.DoorOpen:
				if elevator.Floor == floor {
					if (buttonType == elevio.BT_HallUp && elevator.Direction == elevio.MD_Down) ||
						(buttonType == elevio.BT_HallDown && elevator.Direction == elevio.MD_Up) {

						if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
							elevator.Requests[floor][buttonType] = request.PendingRequest
						}
					}
					doorOpenTimer(elevator)
				} else {
					if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
						elevator.Requests[floor][buttonType] = request.PendingRequest
					}
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
					if (buttonType == elevio.BT_HallUp && elevator.Direction == elevio.MD_Down) ||
						(buttonType == elevio.BT_HallDown && elevator.Direction == elevio.MD_Up) {

						if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
							elevator.Requests[floor][buttonType] = request.PendingRequest
						}
					}
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer(elevator)
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
			switch elevator.State {

			case es.Moving:
				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevator.MotorStopTimer.Stop()

					clearRequestAtFloor(elevator)

					doorOpenTimer(elevator)
					es.SetButtonLights(elevator)

					elevator.State = es.DoorOpen
				} else {
					elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}

			case es.MotorStop:
				elevator.MotorStopTimer.Stop()

				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					clearRequestAtFloor(elevator)
					elevator.MotorStopTimer.Stop()

					doorOpenTimer(elevator)
					es.SetButtonLights(elevator)

					elevator.State = es.DoorOpen
				} else {
					elevator.State = es.Moving
					elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}

			}

		case obstruction := <-obstructionEventCh:
			elevator.Obstruction = obstruction

			if elevator.State == es.DoorOpen {
				elevio.SetDoorOpenLamp(true)
			}
			if !elevator.DoorTimer.HasTimeRemaining() && !elevator.Obstruction {
				onDoorTimeout(elevator)
			}

		case <-elevator.DoorTimer.Timer.C:
			// log("DoorTimer")
			if elevator.Obstruction {
				elevatorUnavailableCh <- true
			} else {
				onDoorTimeout(elevator)
			}

		case <-elevator.MotorStopTimer.Timer.C:
			log("Triggered motor stop timer " + strconv.Itoa(int(elevator.State)))
			switch elevator.State {
			case es.Moving:
				elevator.State = es.MotorStop
				elevatorUnavailableCh <- true
				if !existsRequestsBelow(elevator) && !existsRequestsAbove(elevator) {
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
			// log("Updated Requests")
			elevator.Requests = updatedRequests
			es.SetButtonLights(elevator)

			// Must be here to be able to confirm a new request before moving the elevator (When in idle mode), because it wil choose direction stop if not
			if elevator.State == es.Idle {
				elevator.Direction = chooseDirection(elevator)
				elevio.SetMotorDirection(elevator.Direction)
				elevator.State = es.Moving
				elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
			}
		}
	}
}

func onDoorTimeout(elevator *es.Elevator) {
	floor := elevator.Floor

	// Checks if elevator has a reason to travel in direction decided when floor
	// was reached if uncerviced hall orders exist on the current floor, and services the hall order if not.

	if elevator.State == es.DoorOpen && !elevator.Obstruction {
		if request.IsActive(elevator.Requests[floor][elevio.BT_HallDown]) {
			if !existsRequestsAbove(elevator) {
				var otherStates, _ = network.GetRequestStatesAtIndex(floor, elevio.BT_HallDown)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallDown], otherStates) {
					elevator.Requests[floor][elevio.BT_HallDown] = request.DeleteRequest
				} else {
					log("Could not delete order condition 1, other states:")
					fmt.Printf("%#v", otherStates)
					fmt.Println()
				}
				doorOpenTimer(elevator)
				es.SetButtonLights(elevator)
				elevator.State = es.DoorOpen
				return
			} else {
				var otherStates, _ = network.GetRequestStatesAtIndex(floor, elevio.BT_HallUp)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallUp], otherStates) {
					elevator.Requests[floor][elevio.BT_HallUp] = request.DeleteRequest
				} else {
					log("Could not delete order condition 2, other states:")
					fmt.Printf("%#v", otherStates)
					fmt.Println()
				}
			}
		} else if request.IsActive(elevator.Requests[elevator.Floor][elevio.BT_HallUp]) {
			if !existsRequestsBelow(elevator) {
				var otherStates, _ = network.GetRequestStatesAtIndex(floor, elevio.BT_HallUp)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallUp], otherStates) {
					elevator.Requests[floor][elevio.BT_HallUp] = request.DeleteRequest
				} else {
					log("Could not delete order condition 3, other states:")
					fmt.Printf("%#v", otherStates)
					fmt.Println()
				}
				doorOpenTimer(elevator)
				es.SetButtonLights(elevator)
				elevator.State = es.DoorOpen
				return
			} else {
				var otherStates, _ = network.GetRequestStatesAtIndex(floor, elevio.BT_HallDown)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallDown], otherStates) {
					elevator.Requests[floor][elevio.BT_HallDown] = request.DeleteRequest
				} else {
					log("Could not delete order condition 4, other states:")
					fmt.Printf("%#v", otherStates)
					fmt.Println()
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

func doorOpenTimer(elevator *es.Elevator) {
	elevio.SetDoorOpenLamp(true)
	elevator.DoorTimer.Reset(config.DOOR_OPEN_DURATION)
}
