package fsm

import (
	"Elevator/elevator"
	"Elevator/elevio"
	"time"
)

func RunStateMachine(event_buttonPress <-chan elevio.ButtonEvent, event_floorArrival <-chan int,
					event_obstruction <-chan bool, event_stopButton <-chan bool) {

		elevator := initializeElevator()

		for {
			select{
			case order := <-elevio.ButtonEvent:
				switch elevator.State {
					case MOVING

				}

			case newFloor := <-event_floorArrival:
			}
		}
	elevator.State = elevator.IDLE
	for {
		elevator.Floor = elevio.GetFloor()
		switch elevator.State {
		case elevator.IDLE:
			{
				// Reset the elevator
				elevator.InitializeElevator(elevator)
			}
		case elevator.IDLE_READY:
			if elevator.Direction == elevio.MD_Up {
				elevator.Stop(elevator)
				for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
					if elevator.Requests[floor][elevio.BT_HallUp] || elevator.Requests[floor][elevio.BT_Cab] {
						elevator.GoUp(elevator)
						elevator.State = elevator.MOVING
					}
				}
			} else if elevator.Direction == elevio.MD_Down {
				elevator.Stop(elevator)
				for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
					if elevator.Requests[floor][elevio.BT_HallDown] || elevator.Requests[floor][elevio.BT_Cab] {
						elevator.GoDown(elevator)
						elevator.State = elevator.MOVING
					}
				}
			} else {
				// TODO: Calculate lowest distance away floor and go there
			}
		case elevator.MOVING:
			{
				switch elevator.Direction {
				case elevio.MD_Up:
					if elevio.IsValidFloor(elevator.Floor) {
						if elevator.Requests[elevator.Floor][elevio.BT_HallUp] || elevator.Requests[elevator.Floor][elevio.BT_Cab] {
							elevator.ClearRequestsAtFloor(elevator)
							elevator.Stop(elevator)
							elevator.State = elevator.WAIT
						}
						if elevator.Floor == elevio.NUM_FLOORS-1 {
							elevator.Stop(elevator)
							elevator.State = elevator.WAIT
						}
					}

				case elevio.MD_Down:
					if elevio.IsValidFloor(elevator.Floor) {
						if elevator.Requests[elevator.Floor][elevio.BT_HallDown] || elevator.Requests[elevator.Floor][elevio.BT_Cab] {
							elevator.ClearRequestsAtFloor(elevator)
							elevator.Stop(elevator)
							elevator.State = elevator.WAIT
						}
						if elevator.Floor == 0 {
							elevator.Stop(elevator)
							elevator.State = elevator.WAIT
						}
					}
				default:
					// TODO: Add state and direction dynamically.
					panic("Mismatch between state and direction")
				}
			}
		case elevator.WAIT:
			elevator.Stop(elevator)
			elevator.TryOpenDoor(elevator)
			time.Sleep(elevio.WAIT_DURATION)
			elevator.TryCloseDoor(elevator)

		case elevator.OBSTRUCTED:
			elevator.State = elevator.WAIT
		default:
			elevator.State = elevator.IDLE
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func onRequestButtonPress(button_msg elevio.ButtonEvent, orderCompleteCh chan<- elevio.ButtonEvent, elevator *Elevator) {

	floor := button_msg.Floor
	button_type := button_msg.Button

	switch elevator.state {

	case DoorOpen:
		elevator.requests[floor][button_type] = true
		if elevator.floor == floor {
			clearRequestAtFloor(elevator, orderCompleteCh)
			doorOpenTimer(elevator)
		}

	case Moving:
		elevator.requests[floor][button_type] = true

	case MotorStop:
		elevator.requests[floor][button_type] = true

	case Idle:
		if elevator.floor == floor {
			elevator.requests[floor][button_type] = true
			clearRequestAtFloor(elevator, orderCompleteCh)
			elevio.SetDoorOpenLamp(true)
			// wait some time
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