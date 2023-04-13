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
	peerTxEnable := make(chan bool)

	// Initialize with requests from network (if any)
	go network.NetworkCheck()
	go network.InitSyncReciever(peerTxEnable, requestsUpdate, &elevator.Requests)
	// Ensure we have more than enough time to get requests from network
	time.Sleep(1 * time.Second)
	var cabOrders = network.GetUnionOfLocalCabOrdersFromNetwork()
	// Keep in mind this also includes cab orders from another elevator,
	// to simplify data restructuring.
 	hallOrders, _ := network.GetNewestLocalOrdersFromNetwork()
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for button := 0; button < config.N_BUTTONS; button++ {
			if button == elevio.BT_Cab {
				elevator.Requests[floor][elevio.BT_Cab] = cabOrders[floor]
			} else {
				elevator.Requests[floor][button] = hallOrders[floor][button]
			}
		}
	}
	
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go synchronizer.LocalRequestSynchronization(&elevator, requestsUpdate)
	go network.BroadcastState(&elevator.Floor, &elevator.Direction, &elevator.Obstruction, &elevator.Requests)
	go elevatorFSM.RunStateMachine(&elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, FSM_ElevatorUnavailable, requestsUpdate)
	for {
		time.Sleep(time.Second * 20)
	}
}
