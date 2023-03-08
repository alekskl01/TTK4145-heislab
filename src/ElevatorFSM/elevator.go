package ElevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
	"time"
)

type ElevatorState int

const (
	DoorOpen  ElevatorState = 0
	Moving    ElevatorState = 1
	Idle      ElevatorState = 2
	MotorStop ElevatorState = 3
)

type Elevator struct {
	state          ElevatorState
	floor          int
	direction      elevio.MotorDirection
	requests       [config.N_FLOORS][config.N_BUTTONS]bool
	obstruction    bool
	doorTimer      *time.Timer
	motorStopTimer *time.Timer
}

func clearRequestAtFloor(elev *Elevator, orderComplete chan<- elevio.ButtonEvent) {
	for button := 0; button < config.N_BUTTONS; button++ {
		elev.requests[elev.floor][button] = false
		elevio.SetButtonLamp(elevio.ButtonType(button), elev.floor, false)
		//orderComplete <- elevio.ButtonEvent{Floor: elev.floor, Button: elevio.ButtonType(button)}
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

	//Timers
	elevator.doorTimer = time.NewTimer(config.DOOR_OPEN_DURATION)
	elevator.doorTimer.Stop()
	elevator.motorStopTimer = time.NewTimer(config.MOTOR_STOP_DETECTION_TIME)
	elevator.motorStopTimer.Stop()

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

func setButtonLights(elevator *Elevator) {
	for f := 0; f < config.N_FLOORS; f++ {
		for b := 0; b < config.N_BUTTONS; b++ {
			if elevator.requests[f][b] {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
			}
		}
	}
}
