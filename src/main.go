package main

import (
	"Elevator/ElevatorFSM"
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network"
	"Elevator/network/peers"

	//"fmt"
	"time"
)

func main() {

	// Initiate Networking Channels
	peerUpdateCh := make(chan peers.PeerUpdate)
	stateTxCh := make(chan network.StateMessage)
	stateRxCh := make(chan network.StateMessage)
	actionTxCh := make(chan network.ActionMessage)
	actionRxCh := make(chan network.ActionMessage)

	var networkId = network.GetID()
	network.InitPeerManagement(networkId, peerUpdateCh)
	network.InitStateSynchronizationChannels(networkId, stateTxCh, stateRxCh)
	network.InitActionSynchronizationChannels(networkId, actionTxCh, actionRxCh)

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

	for {
		time.Sleep(time.Second * 20)
	}
}
