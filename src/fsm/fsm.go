package fsm

import (
	"Elevator/elevator"
	"Elevator/elevio"
	"time"
)

func RunStateMachine(system *elevator.Elevator) {
	system.State = elevator.IDLE
	for {
		system.Floor = elevio.GetFloor()
		switch system.State {
		case elevator.IDLE:
			{
				// Reset the system
				elevator.InitializeElevator(system)
			}
		case elevator.IDLE_READY:
			if system.Direction == elevio.MD_Up {
				elevator.Stop(system)
				for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
					if system.Requests[floor][elevio.BT_HallUp] || system.Requests[floor][elevio.BT_Cab] {
						elevator.GoUp(system)
						system.State = elevator.MOVING
					}
				}
			} else if system.Direction == elevio.MD_Down {
				elevator.Stop(system)
				for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
					if system.Requests[floor][elevio.BT_HallDown] || system.Requests[floor][elevio.BT_Cab] {
						elevator.GoDown(system)
						system.State = elevator.MOVING
					}
				}
			} else {
				// TODO: Calculate lowest distance away floor and go there
			}
		case elevator.MOVING:
			{
				switch system.Direction {
				case elevio.MD_Up:
					if elevio.IsValidFloor(system.Floor) {
						if system.Requests[system.Floor][elevio.BT_HallUp] || system.Requests[system.Floor][elevio.BT_Cab] {
							elevator.ClearRequestsAtFloor(system)
							elevator.Stop(system)
							system.State = elevator.WAIT
						}
						if system.Floor == elevio.NUM_FLOORS-1 {
							elevator.Stop(system)
							system.State = elevator.WAIT
						}
					}

				case elevio.MD_Down:
					if elevio.IsValidFloor(system.Floor) {
						if system.Requests[system.Floor][elevio.BT_HallDown] || system.Requests[system.Floor][elevio.BT_Cab] {
							elevator.ClearRequestsAtFloor(system)
							elevator.Stop(system)
							system.State = elevator.WAIT
						}
						if system.Floor == 0 {
							elevator.Stop(system)
							system.State = elevator.WAIT
						}
					}
				default:
					// TODO: Add state and direction dynamically.
					panic("Mismatch between state and direction")
				}
			}
		case elevator.WAIT:
			elevator.Stop(system)
			elevator.TryOpenDoor(system)
			time.Sleep(elevio.WAIT_DURATION)
			elevator.TryCloseDoor(system)

		case elevator.OBSTRUCTED:
			system.State = elevator.WAIT
		default:
			system.State = elevator.IDLE
		}
		time.Sleep(1 * time.Millisecond)
	}
}
