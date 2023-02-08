package elevator

import (
	"Elevator/elevio"
	"fmt"
)

type State int

const (
	IDLE       State = 0
	IDLE_READY       = 1
	MOVING           = 2
	OBSTRUCTED       = 3
	WAIT             = 4
)

type Elevator struct {
	State          State
	Floor          int
	PrevValidFloor int
	Direction      elevio.MotorDirection
	Requests       [elevio.NUM_FLOORS][elevio.NUM_BUTTONS]bool
	IsDoorOpen     bool
}

func ClearRequestsAtFloor(elev *Elevator) {
	for button := 0; button < elevio.NUM_BUTTONS; button++ {
		elev.Requests[elev.Floor][button] = false
		elevio.SetButtonLamp(elevio.ButtonType(button), elev.Floor, false)
	}
}

func ClearAllRequests(elev *Elevator) {
	for floor := 0; floor < elevio.NUM_FLOORS; floor++ {
		for button := 0; button < elevio.NUM_BUTTONS; button++ {
			elev.Requests[elev.Floor][button] = false
			elevio.SetButtonLamp(elevio.ButtonType(button), floor, false)
		}
	}
}

func InitializeElevator(system *Elevator) {
	fmt.Println("PREPARING ELEVATOR...")
	elevio.DefaultInit()
	var floor int = elevio.GetFloor()
	if floor == -1 {
		panic("Tried to initialize elevator at undefined floor")
	}
	system.Direction = elevio.MD_Stop
	system.Floor = elevio.GetFloor()
	// This is a nice template for iterating through the request matrix.
	ClearAllRequests(system)
	system.State = IDLE_READY
}

func Stop(system *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	system.Direction = elevio.MD_Stop
}

func GoUp(system *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Up)
	system.Direction = elevio.MD_Up
}

func GoDown(system *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	system.Direction = elevio.MD_Down
}

func TryOpenDoor(system *Elevator) {
	if elevio.IsValidFloor(system.Floor) {
		elevio.SetDoorOpenLamp(true)
		system.IsDoorOpen = true
	}
}

func TryCloseDoor(system *Elevator) {
	// TODO: Find better way to check if door is obstructed
	if system.IsDoorOpen && !elevio.GetObstruction() {
		system.IsDoorOpen = false
		elevio.SetDoorOpenLamp(false)
	}
}
