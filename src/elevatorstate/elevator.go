package elevatorstate

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/request"
	"fmt"
	"time"
)

type ElevatorState int

const (
	DoorOpen  ElevatorState = 0
	Moving    ElevatorState = 1
	Idle      ElevatorState = 2
	MotorStop ElevatorState = 3
)

type CheckableTimer struct {
	timer *time.Timer
	end   time.Time
}

func NewCheckableTimer(t time.Duration) *CheckableTimer {
	return &CheckableTimer{time.NewTimer(t), time.Now().Add(t)}
}

func (checkable_timer *CheckableTimer) Reset(t time.Duration) {
	checkable_timer.timer.Reset(t)
	checkable_timer.end = time.Now().Add(t)
}

func (checkable_timer *CheckableTimer) Stop() {
	checkable_timer.timer.Stop()
}

func (checkable_timer *CheckableTimer) hasTimeRemaining() bool {
	time_left := time.Until(checkable_timer.end)
	fmt.Println(time_left)
	return (time_left > 0)
}

type Elevator struct {
	State          ElevatorState
	Floor          int
	Direction      elevio.MotorDirection
	Requests       [config.N_FLOORS][config.N_BUTTONS]request.RequestState
	Obstruction    bool
	DoorTimer      CheckableTimer
	MotorStopTimer CheckableTimer
}

func InitializeElevator() Elevator {
	elevator := new(Elevator)
	elevator.Floor = -1
	elevator.Direction = elevio.MD_Stop
	elevator.State = Idle
	elevator.Obstruction = false

	//Timers
	elevator.DoorTimer = *NewCheckableTimer(config.DOOR_OPEN_DURATION)
	elevator.DoorTimer.Stop()
	elevator.MotorStopTimer = *NewCheckableTimer(config.MOTOR_STOP_DETECTION_TIME)
	elevator.MotorStopTimer.Stop()

	//Make sure elevator is not between floors
	elevator.Direction = elevio.MD_Down
	elevio.SetMotorDirection(elevio.MD_Down)
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
			if request.ShouldActivateButtonLight(elevator.Requests[f][b]) {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
			}
		}
	}
}
