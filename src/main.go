package main

import (
	"Elevator/config"
	"Elevator/elevatorstate"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/request"
	"Elevator/statemanager"
	"Elevator/synchronizer"
	"flag"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting elevator")
	port := flag.Int("port", 0, "")
	flag.Parse()
	if *port != 0 {
		config.Port = *port
	}
	network.LocalID = network.GetID()
	elevio.Init()
	elevator := elevatorstate.InitializeElevator()
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	FSM_ElevatorUnavailable := make(chan bool, config.FSM_CHANNEL_BUFFER_SIZE)
	requestsUpdate := make(chan [config.N_FLOORS][config.N_BUTTONS]request.RequestState)
	peerTxEnable := make(chan bool)

	// Initialize with requests from network (if any)
	go network.NetworkCheck(&elevator.Requests)
	go network.PeerUpdateReciever(peerTxEnable, requestsUpdate, &elevator.Requests)
	go network.SyncReciever()
	// Ensure we have more than enough time to get requests from network
	time.Sleep(config.SIGNIFICANT_DELAY)

	// Get any cab order we had previously from the network
	var cabOrders = network.GetLocalCabOrdersFromNetwork()
	for floor := 0; floor < config.N_FLOORS; floor++ {
		elevator.Requests[floor][elevio.BT_Cab] = cabOrders[floor]
	}
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go synchronizer.LocalRequestSynchronization(&elevator, requestsUpdate)
	statemanager.InitCheapestRequests()
	go synchronizer.UpdateCheapestRequests(&elevator.Floor, &elevator.Direction, &elevator.State, &elevator.Obstruction, &elevator.Requests)
	go network.BroadcastState(&elevator.Floor, &elevator.Direction, &elevator.State, &elevator.Obstruction, &elevator.Requests)
	go statemanager.RunStateMachine(&elevator, drv_buttons, drv_floors, drv_obstr, drv_stop, FSM_ElevatorUnavailable, requestsUpdate)
	for {
		time.Sleep(time.Second * 20)
	}
}
