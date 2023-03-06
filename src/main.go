package main

import (

	"Elevator/elevio"
	"Elevator/ElevatorFSM"
	"Elevator/config"
)

func main() {

	elevio.DefaultInit()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	FSM_OrderComplete := make(chan elevio.ButtonEvent, config.FSM_CHANNEL_BUFFER_SIZE)
	FSM_ElevatorUnavailable := make(chan bool, config.FSM_CHANNEL_BUFFER_SIZE)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go ElevatorFSM.RunStateMachine(drv_buttons, drv_floors, drv_obstr, drv_stop, FSM_ElevatorUnavailable, FSM_OrderComplete)
}
