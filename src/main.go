package main

import (
	"Elevator/config"
	"Elevator/elevatorFSM"
	"Elevator/elevio"
	"Elevator/network"
	"time"
)

func main() {

	elevio.DefaultInit()
	elevator := elevatorFSM.InitializeElevator()
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	FSM_ElevatorUnavailable := make(chan bool, config.FSM_CHANNEL_BUFFER_SIZE)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go network.BroadcastState(&elevator.Floor, &elevator.Direction, &elevator.Requests)
	go network.InitSyncReciever()
	go elevatorFSM.RunStateMachine(&elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, FSM_ElevatorUnavailable)

	for {
		time.Sleep(time.Second * 20)
	}
}
