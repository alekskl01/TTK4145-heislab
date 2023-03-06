package ElevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
)

func RunStateMachine(event_buttonPress <-chan elevio.ButtonEvent, event_floorArrival <-chan int,
	event_obstruction <-chan bool, event_stopButton <-chan bool, ch_elevatorUnavailable chan<- bool, ch_orderComplete chan<- elevio.ButtonEvent) {

	elevator := InitializeElevator()

	for {
		select {
		case order := <-event_buttonPress:
			floor := order.Floor
			button_type := order.Button

			switch elevator.state {

			case DoorOpen:
				elevator.requests[floor][button_type] = true
				if elevator.floor == floor {
					clearRequestAtFloor(&elevator, ch_orderComplete)
					doorOpenTimer(&elevator)
				}

			case Moving:
				elevator.requests[floor][button_type] = true

			case MotorStop:
				elevator.requests[floor][button_type] = true

			case Idle:
				if elevator.floor == floor {
					elevator.requests[floor][button_type] = true
					clearRequestAtFloor(&elevator, ch_orderComplete)
					elevio.SetDoorOpenLamp(true)
					doorOpenTimer(&elevator)
					elevator.state = DoorOpen
				} else {
					elevator.requests[floor][button_type] = true
					elevator.direction = chooseDirection(elevator)
					elevio.SetMotorDirection(elevator.direction)
					elevator.state = Moving
					elevator.motorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}
			}
			setCabLights(&elevator)

		case newFloor := <-event_floorArrival:
			elevator.floor = newFloor

			elevio.SetFloorIndicator(newFloor)
			switch elevator.state {

			case Moving:
				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					clearRequestAtFloor(&elevator, ch_orderComplete)
					elevator.motorStopTimer.Stop()

					doorOpenTimer(&elevator)
					setCabLights(&elevator)

					elevator.state = DoorOpen
				} else {
					elevator.motorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}

			case MotorStop:
				elevator.motorStopTimer.Stop()

				if shouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					clearRequestAtFloor(&elevator, ch_orderComplete)
					elevator.motorStopTimer.Stop()

					doorOpenTimer(&elevator)
					setCabLights(&elevator)

					elevator.state = DoorOpen
				} else {
					elevator.state = Moving
					elevator.motorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
				}

			}

		case obstruction := <-event_obstruction:
			elevator.obstruction = obstruction

			if elevator.state == DoorOpen {
				elevio.SetDoorOpenLamp(true)
			}

			//onDoorTimeout(&elevator)

		case <-elevator.motorStopTimer.C:
			switch elevator.state {
			case Moving:
				elevator.state = MotorStop
				ch_elevatorUnavailable <- true
				if !existsRequestsBelow(elevator) && !existsRequestsAbove(elevator) {
					elevator.requests[elevator.floor+int(elevator.direction)][elevio.BT_Cab] = true
				}
				elevator.motorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)

			case MotorStop:
				elevator.direction = chooseDirection(elevator)
				elevio.SetMotorDirection(elevator.direction)
				elevator.motorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
			}
		}
	}
}

func onRequestButtonPress(button_msg elevio.ButtonEvent, elevator *Elevator, ch_orderComplete chan<- elevio.ButtonEvent) {

	floor := button_msg.Floor
	button_type := button_msg.Button

	switch elevator.state {

	case DoorOpen:
		elevator.requests[floor][button_type] = true
		if elevator.floor == floor {
			clearRequestAtFloor(elevator, ch_orderComplete)
			doorOpenTimer(elevator)
		}

	case Moving:
		elevator.requests[floor][button_type] = true

	case MotorStop:
		elevator.requests[floor][button_type] = true

	case Idle:
		if elevator.floor == floor {
			elevator.requests[floor][button_type] = true
			clearRequestAtFloor(elevator, ch_orderComplete)
			elevio.SetDoorOpenLamp(true)
			doorOpenTimer(elevator)
			elevator.state = DoorOpen
		} else {
			elevator.requests[floor][button_type] = true
			elevator.direction = chooseDirection(*elevator)
			elevio.SetMotorDirection(elevator.direction)
			elevator.state = Moving
			elevator.motorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
		}
	}
}

func onDoorTimeout(elevator *Elevator) {
	if elevator.state == DoorOpen && !elevator.obstruction {
		elevio.SetDoorOpenLamp(false)
		elevator.direction = chooseDirection(*elevator)
		elevio.SetMotorDirection(elevator.direction)

		if elevator.direction == elevio.MD_Stop {
			elevator.state = Idle
		} else {
			elevator.state = Moving
			elevator.motorStopTimer.Reset(config.MOTOR_STOP_DETECTION_TIME)
		}
	}
}

func doorOpenTimer(elevator *Elevator) {
	const doorOpenTime = config.DOOR_OPEN_DURATION
	elevio.SetDoorOpenLamp(true)
	elevator.doorTimer.Reset(doorOpenTime)
}
