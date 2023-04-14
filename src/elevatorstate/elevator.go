// This file manages and keeps track of the local elevator on a low level,
// and servers as a middleman between the main state machine and elevio
package elevatorstate

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/request"
	"time"
)

type ElevatorState int

const (
	DoorOpen  ElevatorState = 0
	Moving    ElevatorState = 1
	Idle      ElevatorState = 2
	MotorStop ElevatorState = 3
)

// Expanded timer object that is needed to check remaining time left.
type CheckableTimer struct {
	timer *time.Timer
	end   time.Time
}

func createNewCheckableTimer(t time.Duration) *CheckableTimer {
	return &CheckableTimer{time.NewTimer(t), time.Now().Add(t)}
}

func (checkableTimer *CheckableTimer) reset(t time.Duration) {
	checkableTimer.timer.Reset(t)
	checkableTimer.end = time.Now().Add(t)
}

func (checkableTimer *CheckableTimer) stop() {
	checkableTimer.timer.Stop()
}

func (checkable_timer *CheckableTimer) hasTimeRemaining() bool {
	time_left := time.Until(checkable_timer.end)
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
	elevator.DoorTimer = *createNewCheckableTimer(config.DOOR_OPEN_DURATION)
	elevator.DoorTimer.stop()
	elevator.MotorStopTimer = *createNewCheckableTimer(config.MOTOR_STOP_DETECTION_TIME)
	elevator.MotorStopTimer.stop()

	//Make sure elevator is not between floors
	elevator.Direction = elevio.MD_Down
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.State = Moving

	return *elevator
}

func setButtonLights(elevator *Elevator) {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if request.ShouldActivateButtonLight(elevator.Requests[floor][button]) {
				elevio.SetButtonLamp(elevio.ButtonType(button), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(button), floor, false)
			}
		}
	}
}
