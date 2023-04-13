package elevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
	"fmt"
)

var CheapestRequests [config.N_FLOORS][config.N_BUTTONS]bool

func InitCheapestRequests() {
	cheapestRequests := make([][]bool, config.N_FLOORS)
	for i := range cheapestRequests {
		cheapestRequests[i] = make([]bool, config.N_BUTTONS)
	}

	// Cab orders are always cheapest for us to take
	for floor := 0; floor < config.N_FLOORS; floor++ {
		cheapestRequests[floor][0] = true
	}
}

func RunStateMachine(elevator *Elevator, event_buttonPress <-chan elevio.ButtonEvent, event_floorArrival <-chan int,
	event_obstruction <-chan bool, event_stopButton <-chan bool, ch_elevatorUnavailable chan<- bool, event_requestsUpdated <-chan [config.N_FLOORS][config.N_BUTTONS]request.RequestState) {

	for {
		select {
		case order := <-event_buttonPress:
			floor := order.Floor
			button_type := order.Button
			var otherStates = network.GetRequestStatesAtIndex(floor, button_type)

			// Prevents taking hall orders if we are not connected to the network
			if len(otherStates) == 0 && button_type != elevio.BT_Cab { 
				break
			}
			
			switch elevator.State {
			case DoorOpen:
				if elevator.Floor == floor {
					if (button_type == elevio.BT_HallUp && elevator.Direction == elevio.MD_Down) ||
						(button_type == elevio.BT_HallDown && elevator.Direction == elevio.MD_Up) {

						if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][button_type], otherStates) {
							elevator.Requests[floor][button_type] = request.PendingRequest
						}
					}
					doorOpenTimer(elevator)
				} else {
					if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][button_type], otherStates) {
						elevator.Requests[floor][button_type] = request.PendingRequest
					}
				}

			case Moving:
				if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][button_type], otherStates) {
					elevator.Requests[floor][button_type] = request.PendingRequest
				}

			case MotorStop:
				if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][button_type], otherStates) {
					elevator.Requests[floor][button_type] = request.PendingRequest
				}

			case Idle:
				if elevator.Floor == floor {
					if (button_type == elevio.BT_HallUp && elevator.Direction == elevio.MD_Down) ||
						(button_type == elevio.BT_HallDown && elevator.Direction == elevio.MD_Up) {

						if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][button_type], otherStates) {
							elevator.Requests[floor][button_type] = request.PendingRequest
						}
					}
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer(elevator)
					elevator.State = DoorOpen
				} else {
					if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[floor][button_type], otherStates) {
						elevator.Requests[floor][button_type] = request.PendingRequest
					}
				}
			}
			setButtonLights(elevator)

		case newFloor := <-event_floorArrival:
			elevator.Floor = newFloor

			elevio.SetFloorIndicator(newFloor)
			switch elevator.State {

			case Moving:
				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevator.MotorStopTimer.Stop()

					clearRequestAtFloor(elevator)

					doorOpenTimer(elevator)
					setButtonLights(elevator)

					elevator.State = DoorOpen
				} else {
					elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}

			case MotorStop:
				elevator.MotorStopTimer.Stop()

				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					clearRequestAtFloor(elevator)
					elevator.MotorStopTimer.Stop()

					doorOpenTimer(elevator)
					setButtonLights(elevator)

					elevator.State = DoorOpen
				} else {
					elevator.State = Moving
					elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}

			}

		case obstruction := <-event_obstruction:
			elevator.Obstruction = obstruction

			if elevator.State == DoorOpen {
				elevio.SetDoorOpenLamp(true)
			}
			if !elevator.DoorTimer.hasTimeRemaining() && !elevator.Obstruction {
				onDoorTimeout(elevator)
			}

		case <-elevator.DoorTimer.timer.C:
			Log("DoorTimer")
			if elevator.Obstruction {
				ch_elevatorUnavailable <- true
			} else {
				onDoorTimeout(elevator)
			}

		case <-elevator.MotorStopTimer.timer.C:
			Log("Triggered motor stop timer")
			switch elevator.State {
			case Moving:
				elevator.State = MotorStop
				ch_elevatorUnavailable <- true
				if !existsRequestsBelow(elevator) && !existsRequestsAbove(elevator) {
					var otherStates = network.GetRequestStatesAtIndex(elevator.Floor+int(elevator.Direction), elevio.BT_Cab)
					if request.OrderStatesEqualTo(request.NoRequest, elevator.Requests[elevator.Floor+int(elevator.Direction)][elevio.BT_Cab], otherStates) {
						elevator.Requests[elevator.Floor+int(elevator.Direction)][elevio.BT_Cab] = request.PendingRequest
					}
				}
				elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)

			case MotorStop:
				elevator.Direction = chooseDirection(elevator)
				elevio.SetMotorDirection(elevator.Direction)
				elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
			}
		case updated_requests := <-event_requestsUpdated:
			Log("Updated Requests")
			elevator.Requests = updated_requests
			setButtonLights(elevator)

			// Must be here to be able to confirm a new request before moving the elevator (When in idle mode), because it wil choose direction stop if not
			if elevator.State == Idle {
				elevator.Direction = chooseDirection(elevator)
				elevio.SetMotorDirection(elevator.Direction)
				elevator.State = Moving
				elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
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
					Log("Could not delete order 1")
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
					Log("Could not delete order 2")
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
					Log("Could not delete order 3")
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
					Log("Could not delete order 4")
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
			elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
		}
	}
}

func doorOpenTimer(elevator *Elevator) {
	const doorOpenTime = config.DOOR_OPEN_DURATION
	elevio.SetDoorOpenLamp(true)
	elevator.DoorTimer.Reset(doorOpenTime)
}

func Log(text string) {
	fmt.Println("FSM: " + text)
}
