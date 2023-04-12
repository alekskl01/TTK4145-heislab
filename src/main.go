package main

import (
	"Elevator/config"
	"Elevator/elevatorFSM"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
	"Elevator/synchronizer"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting elevator")
	elevio.DefaultInit()
	elevator := elevatorFSM.InitializeElevator()
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	FSM_ElevatorUnavailable := make(chan bool, config.FSM_CHANNEL_BUFFER_SIZE)
	requestsUpdate := make(chan [config.N_FLOORS][config.N_BUTTONS]request.RequestState)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	//go synchronizer.GlobalRequestSynchronization()
	go synchronizer.LocalRequestSynchronization(&elevator, requestsUpdate)
	go network.BroadcastState(&elevator.Floor, &elevator.Direction, &elevator.Obstruction, &elevator.Requests)
	go network.InitSyncReciever()
	go elevatorFSM.RunStateMachine(&elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, FSM_ElevatorUnavailable, requestsUpdate)
	for {
		time.Sleep(time.Second * 20)
	}
}
