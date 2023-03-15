package ElevatorFSM

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network"
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
	State          ElevatorState
	Floor          int
	Direction      elevio.MotorDirection
	Requests       [config.N_FLOORS][config.N_BUTTONS]bool
	Obstruction    bool
	DoorTimer      *time.Timer
	MotorStopTimer *time.Timer
}

func InitializeElevator() Elevator {
	elevator := new(Elevator)
	elevator.Floor = -1
	elevator.Direction = elevio.MD_Stop
	elevator.State = Idle
	elevator.Obstruction = false

	//Timers
	elevator.DoorTimer = time.NewTimer(config.DOOR_OPEN_DURATION)
	elevator.DoorTimer.Stop()
	elevator.MotorStopTimer = time.NewTimer(config.MOTOR_STOP_DETECTION_TIME)
	elevator.MotorStopTimer.Stop()

	//Make sure elevator is not between floors
	elevator.Direction = elevio.MD_Down
	elevio.SetMotorDirection(elevator.Direction)
	elevator.State = Moving

	return *elevator
}

func Stop(elevator *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevator.Direction = elevio.MD_Stop
}

func GoUp(elevator *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Up)
	elevator.Direction = elevio.MD_Up
}

func GoDown(elevator *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Direction = elevio.MD_Down
}

func setButtonLights(elevator *Elevator) {
	for f := 0; f < config.N_FLOORS; f++ {
		for b := 0; b < config.N_BUTTONS; b++ {
			if elevator.Requests[f][b] {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
			}
		}
	}
}

func BroadcastState(elevator *Elevator, stateCh chan<- network.StateMessage) {
	for {
		stateCh <- network.StateMessage{elevator.Floor, elevator.Direction}
		time.Sleep(time.Millisecond * 200)
	}
}
