// This file contains the main state machine that governs the behaviour of the local elevator
package elevatorstate

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
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

func RunStateMachine(elevator *Elevator, buttonPressEventCh <-chan elevio.ButtonEvent, floorArrivalEventCh <-chan int,
	obstructionEventCh <-chan bool, stopButtonEventCh <-chan bool, elevatorUnavailableCh chan<- bool, requestsUpdatedCh <-chan [config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	for {
		select {
		case order := <-buttonPressEventCh:
			floor := order.Floor
			buttonType := order.Button
			var otherStates = network.GetRequestStatesAtIndex(floor, buttonType)

			// Prevents taking hall orders if we are not connected to the network
			if len(otherStates) == 0 && buttonType != elevio.BT_Cab {
				break
			}

			switch elevator.State {
			case DoorOpen:
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

			case Moving:
				if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
					elevator.Requests[floor][buttonType] = request.PendingRequest
				}

			case MotorStop:
				if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
					elevator.Requests[floor][buttonType] = request.PendingRequest
				}

			case Idle:
				if elevator.Floor == floor {
					if (buttonType == elevio.BT_HallUp && elevator.Direction == elevio.MD_Down) ||
						(buttonType == elevio.BT_HallDown && elevator.Direction == elevio.MD_Up) {

						if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
							elevator.Requests[floor][buttonType] = request.PendingRequest
						}
					}
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer(elevator)
					elevator.State = DoorOpen
				} else {
					if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][buttonType], otherStates) {
						elevator.Requests[floor][buttonType] = request.PendingRequest
					}
				}
			}
			setButtonLights(elevator)

		case newFloor := <-floorArrivalEventCh:
			elevator.Floor = newFloor

			elevio.SetFloorIndicator(newFloor)
			switch elevator.State {

			case Moving:
				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevator.MotorStopTimer.stop()

					clearRequestAtFloor(elevator)

					doorOpenTimer(elevator)
					setButtonLights(elevator)

					elevator.State = DoorOpen
				} else {
					elevator.MotorStopTimer.reset(config.MOTOR_STOP_DETECTION_TIME)
				}

			case MotorStop:
				elevator.MotorStopTimer.stop()

				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					clearRequestAtFloor(elevator)
					elevator.MotorStopTimer.stop()

					doorOpenTimer(elevator)
					setButtonLights(elevator)

					elevator.State = DoorOpen
				} else {
					elevator.State = Moving
					elevator.MotorStopTimer.reset(config.MOTOR_STOP_DETECTION_TIME)
				}

			}

		case obstruction := <-obstructionEventCh:
			elevator.Obstruction = obstruction

			if elevator.State == DoorOpen {
				elevio.SetDoorOpenLamp(true)
			}
			if !elevator.DoorTimer.hasTimeRemaining() && !elevator.Obstruction {
				onDoorTimeout(elevator)
			}

		case <-elevator.DoorTimer.timer.C:
			// log("DoorTimer")
			if elevator.Obstruction {
				elevatorUnavailableCh <- true
			} else {
				onDoorTimeout(elevator)
			}

		case <-elevator.MotorStopTimer.timer.C:
			log("Triggered motor stop timer " + strconv.Itoa(int(elevator.State)))
			switch elevator.State {
			case Moving:
				elevator.State = MotorStop
				elevatorUnavailableCh <- true
				if !existsRequestsBelow(elevator) && !existsRequestsAbove(elevator) {
					var otherStates = network.GetRequestStatesAtIndex(elevator.Floor+int(elevator.Direction), elevio.BT_Cab)
					if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[elevator.Floor+int(elevator.Direction)][elevio.BT_Cab], otherStates) {
						elevator.Requests[elevator.Floor+int(elevator.Direction)][elevio.BT_Cab] = request.PendingRequest
					}
				}
				elevator.MotorStopTimer.reset(config.MOTOR_STOP_DETECTION_TIME)

			case MotorStop:
				elevator.Direction = chooseDirection(elevator)
				elevio.SetMotorDirection(elevator.Direction)
				elevator.MotorStopTimer.reset(config.MOTOR_STOP_DETECTION_TIME)
			}
		case updatedRequests := <-requestsUpdatedCh:
			// log("Updated Requests")
			elevator.Requests = updatedRequests
			setButtonLights(elevator)

			// Must be here to be able to confirm a new request before moving the elevator (When in idle mode), because it wil choose direction stop if not
			if elevator.State == Idle {
				elevator.Direction = chooseDirection(elevator)
				elevio.SetMotorDirection(elevator.Direction)
				elevator.State = Moving
				elevator.MotorStopTimer.reset(config.MOTOR_STOP_DETECTION_TIME)
			}
		}
	}
}

func onDoorTimeout(elevator *Elevator) {
	floor := elevator.Floor

	// Checks if elevator has a reason to travel in direction decided when floor
	// was reached if uncerviced hall orders exist on the current floor, and services the hall order if not.

	if elevator.State == DoorOpen && !elevator.Obstruction {
		if request.IsActive(elevator.Requests[floor][elevio.BT_HallDown]) {
			if !existsRequestsAbove(elevator) {
				var otherStates = network.GetRequestStatesAtIndex(floor, elevio.BT_HallDown)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallDown], otherStates) {
					elevator.Requests[floor][elevio.BT_HallDown] = request.DeleteRequest
				} else {
					log("Could not delete order condition 1, other states:")
					fmt.Printf("%#v", otherStates)
					fmt.Println()
				}
				doorOpenTimer(elevator)
				setButtonLights(elevator)
				elevator.State = DoorOpen
				return
			} else {
				var otherStates = network.GetRequestStatesAtIndex(floor, elevio.BT_HallUp)
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
				var otherStates = network.GetRequestStatesAtIndex(floor, elevio.BT_HallUp)
				if request.OrderStatesEqualTo(request.ActiveRequest, elevator.Requests[floor][elevio.BT_HallUp], otherStates) {
					elevator.Requests[floor][elevio.BT_HallUp] = request.DeleteRequest
				} else {
					log("Could not delete order condition 3, other states:")
					fmt.Printf("%#v", otherStates)
					fmt.Println()
				}
				doorOpenTimer(elevator)
				setButtonLights(elevator)
				elevator.State = DoorOpen
				return
			} else {
				var otherStates = network.GetRequestStatesAtIndex(floor, elevio.BT_HallDown)
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
			elevator.State = Idle
		} else {
			elevator.State = Moving
			elevator.MotorStopTimer.reset(config.MOTOR_STOP_DETECTION_TIME)
		}
	}
}

func doorOpenTimer(elevator *Elevator) {
	elevio.SetDoorOpenLamp(true)
	elevator.DoorTimer.reset(config.DOOR_OPEN_DURATION)
}
