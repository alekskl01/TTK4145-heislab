package ElevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network"
)

func RunStateMachine(event_buttonPress <-chan elevio.ButtonEvent, event_floorArrival <-chan int,
	event_obstruction <-chan bool, event_stopButton <-chan bool, ch_elevatorUnavailable chan<- bool, actionTxCh chan<- network.ActionMessage) {

	elevator := InitializeElevator()

	for {
		select {
		case order := <-event_buttonPress:
			floor := order.Floor
			button_type := order.Button

			switch elevator.State {
			case DoorOpen:
				// TODO: check that all counters are in the right and same state before adding or removing orders
				elevator.Requests[floor][button_type] = true
				if elevator.Floor == floor {
					clearRequestAtFloor(&elevator, actionTxCh)
					doorOpenTimer(&elevator)
				}

			case Moving:
				elevator.Requests[floor][button_type] = true

			case MotorStop:
				elevator.Requests[floor][button_type] = true

			case Idle:
				if elevator.Floor == floor {
					elevator.Requests[floor][button_type] = true
					clearRequestAtFloor(&elevator, actionTxCh)
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer(&elevator)
					elevator.State = DoorOpen
				} else {
					elevator.Requests[floor][button_type] = true
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
					elevator.Requests[elevator.Floor+int(elevator.Direction)][elevio.BT_Cab] = true
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
				elevator.Requests[floor][elevio.BT_HallDown] = false
				doorOpenTimer(elevator)
				setButtonLights(elevator)
				elevator.State = DoorOpen
			}
		} else if elevator.Requests[elevator.Floor][elevio.BT_HallUp] {
			if !existsRequestsBelow(*elevator) {
				elevator.Requests[floor][elevio.BT_HallUp] = false
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
