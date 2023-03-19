package elevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
)

func RunStateMachine(elevator *Elevator, event_buttonPress <-chan elevio.ButtonEvent, event_floorArrival <-chan int,
	event_obstruction <-chan bool, event_stopButton <-chan bool, ch_elevatorUnavailable chan<- bool, actionTxCh chan<- network.ActionMessage) {

	for {
		select {
		case order := <-event_buttonPress:
			floor := order.Floor
			button_type := order.Button

			switch elevator.State {
			case DoorOpen:
				if orderStatesEqualTo(NoRequest, elevator.Requests[floor][button_type], otherStates) {
					elevator.Requests[floor][button_type] = PendingRequest
				}

				if elevator.Floor == floor {
					clearRequestAtFloor(&elevator, actionTxCh)
					doorOpenTimer(&elevator)
				}

			case Moving:
				if orderStatesEqualTo(NoRequest, elevator.Requests[floor][button_type], otherStates) {
					elevator.Requests[floor][button_type] = PendingRequest
				}

			case MotorStop:
				if orderStatesEqualTo(NoRequest, elevator.Requests[floor][button_type], otherStates) {
					elevator.Requests[floor][button_type] = PendingRequest
				}

			case Idle:
				if elevator.Floor == floor {
					if orderStatesEqualTo(NoRequest, elevator.Requests[floor][button_type], otherStates) {
						elevator.Requests[floor][button_type] = PendingRequest
					}
					clearRequestAtFloor(&elevator, actionTxCh)
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer(&elevator)
					elevator.State = DoorOpen
				} else {
					if orderStatesEqualTo(NoRequest, elevator.Requests[floor][button_type], otherStates) {
						elevator.Requests[floor][button_type] = PendingRequest
					}
					elevator.Direction = chooseDirection(elevator)
					elevio.SetMotorDirection(elevator.Direction)
					elevator.State = Moving
					elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}
			}
			if elevator.Requests[floor][button_type] {
				actionTxCh <- network.ActionMessage{floor, network.NewRequest}
			}
			setButtonLights(&elevator)

		case newFloor := <-event_floorArrival:
			elevator.Floor = newFloor

			elevio.SetFloorIndicator(newFloor)
			switch elevator.State {

			case Moving:
				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)

					clearRequestAtFloor(&elevator, actionTxCh)
					elevator.MotorStopTimer.Stop()

					doorOpenTimer(&elevator)
					setButtonLights(&elevator)

					elevator.State = DoorOpen
				} else {
					elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}

			case MotorStop:
				elevator.MotorStopTimer.Stop()

				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					clearRequestAtFloor(&elevator, actionTxCh)
					elevator.MotorStopTimer.Stop()

					doorOpenTimer(&elevator)
					setButtonLights(&elevator)

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

			onDoorTimeout(&elevator)

		case <-elevator.DoorTimer.C:
			if elevator.Obstruction {
				ch_elevatorUnavailable <- true
			} else {
				onDoorTimeout(&elevator)
			}

		case <-elevator.MotorStopTimer.C:
			switch elevator.State {
			case Moving:
				elevator.State = MotorStop
				ch_elevatorUnavailable <- true
				if !existsRequestsBelow(elevator) && !existsRequestsAbove(elevator) {
					if orderStatesEqualTo(NoRequest, elevator.Requests[elevator.Floor+int(elevator.Direction)][elevio.BT_Cab], otherStates) {
						elevator.Requests[elevator.Floor+int(elevator.Direction)][elevio.BT_Cab] = PendingRequest
					}
				}
				elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)

			case MotorStop:
				elevator.Direction = chooseDirection(elevator)
				elevio.SetMotorDirection(elevator.Direction)
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
		if elevator.Requests[floor][elevio.BT_HallDown] {
			if !existsRequestsAbove(*elevator) {
				if orderStatesEqualTo(ActiveRequest, elevator.Requests[floor][elevio.BT_HallDown], otherStates) {
					elevator.Requests[floor][elevio.BT_HallDown] = DeleteRequest
				}
				doorOpenTimer(elevator)
				setButtonLights(elevator)
				elevator.State = DoorOpen
			}
		} else if elevator.Requests[elevator.Floor][elevio.BT_HallUp] {
			if !existsRequestsBelow(*elevator) {
				if orderStatesEqualTo(ActiveRequest, elevator.Requests[floor][elevio.BT_HallUp], otherStates) {
					elevator.Requests[floor][elevio.BT_HallUp] = DeleteRequest
				}
				doorOpenTimer(elevator)
				setButtonLights(elevator)
				elevator.State = DoorOpen
			}
		} else {
			elevio.SetDoorOpenLamp(false)
			elevator.Direction = chooseDirection(*elevator)
			elevio.SetMotorDirection(elevator.Direction)

			if elevator.Direction == elevio.MD_Stop {
				elevator.State = Idle
			} else {
				elevator.State = Moving
				elevator.MotorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
			}
		}
	}
}

func doorOpenTimer(elevator *Elevator) {
	const doorOpenTime = config.DOOR_OPEN_DURATION
	elevio.SetDoorOpenLamp(true)
	elevator.DoorTimer.Reset(doorOpenTime)
}
