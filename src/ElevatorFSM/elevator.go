package ElevatorFSM

import (
	"Elevator/elevio"
	"Elevator/config"
	//"fmt"
)

type ElevatorState int

const (
	DoorOpen ElevatorState = 0
	Moving ElevatorState = 1
	Idle ElevatorState = 2
	MotorStop ElevatorState = 3
)

type Elevator struct {
	state          ElevatorState
	floor          int
	direction      elevio.MotorDirection
	requests       [config.N_FLOORS][config.N_BUTTONS]bool
	obstruction     bool
}

func clearRequestAtFloor(elev *Elevator) {
	for button := 0; button < config.N_BUTTONS; button++ {
		elev.requests[elev.floor][button] = false
		elevio.SetButtonLamp(elevio.ButtonType(button), elev.floor, false)
	}
}

func clearAllRequests(elev *Elevator) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			elev.requests[elev.floor][button] = false
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, false)
		}
	}
}

func InitializeElevator() Elevator {
	elevator := new(Elevator)
	elevator.floor = -1
	elevator.direction = elevio.MD_Stop
	elevator.state = Idle
	elevator.obstruction = false

	//Make sure elevator is not between floors
	elevator.direction = elevio.MD_Down
	elevio.SetMotorDirection(elevator.direction)
	elevator.state = Moving

	return *elevator
}

func Stop(elevator *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevator.direction = elevio.MD_Stop
}

func GoUp(elevator *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Up)
	elevator.direction = elevio.MD_Up
}

func GoDown(elevator *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.direction = elevio.MD_Down
}


