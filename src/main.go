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

	// Allows port to be specified when running the application using -port 12345
	port := flag.Int("port", 0, "")
	flag.Parse()
	if *port != 0 {
		config.Port = *port
	}

	network.LocalID = network.GetID()
	elevio.Init()
	elevator := elevatorstate.InitializeElevator()

	buttonDriverCh := make(chan elevio.ButtonEvent)
	floorDriverCh := make(chan int)
	obstructionDriverCh := make(chan bool)
	stopDriverCh := make(chan bool)
	requestsUpdateCh := make(chan [config.N_FLOORS][config.N_BUTTONS]request.RequestState)
	peerTxEnableCh := make(chan bool)

	// Prints network synchronization information when the stop button is pressed.
	go network.NetworkCheck(&elevator.Requests, stopDriverCh)

	// Start recieving networking synchronization states and connection information.
	go network.PeerUpdateReciever(peerTxEnableCh, requestsUpdateCh, &elevator.Requests)
	go network.SyncReciever()
	// Ensure we have more than enough time to get requests from network.
	time.Sleep(config.SIGNIFICANT_DELAY)

	// Get any cab order this node had prior to restart from the network.
	var cabOrders = network.GetLocalCabOrdersFromNetwork()
	for floor := 0; floor < config.N_FLOORS; floor++ {
		elevator.Requests[floor][elevio.BT_Cab] = cabOrders[floor]
	}

	// Start taking inputs from the elevator system.
	go elevio.PollButtons(buttonDriverCh)
	go elevio.PollFloorSensor(floorDriverCh)
	go elevio.PollObstructionSwitch(obstructionDriverCh)
	go elevio.PollStopButton(stopDriverCh)

	// Start the lifecycles of local requests
	go synchronizer.LocalRequestSynchronization(&elevator, requestsUpdateCh)

	// Information from the elevator system and regarding local requests is available and can be broadcasted.
	go network.BroadcastState(&elevator.Floor, &elevator.Direction, &elevator.State, &elevator.Obstruction, &elevator.Requests)

	// Make information about cheapest requests for us to take available to the main state machine.
	statemanager.InitSelfAssignedRequests()
	go synchronizer.UpdateSelfAssignedRequests(&elevator.Floor, &elevator.Direction, &elevator.State, &elevator.Obstruction, &elevator.Requests)
	// This begins the main elevator operation
	go statemanager.RunStateMachine(&elevator, buttonDriverCh, floorDriverCh, obstructionDriverCh, requestsUpdateCh)
	for {
		time.Sleep(time.Second * 20)
	}
}
