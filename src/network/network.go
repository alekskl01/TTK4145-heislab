package network

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network/bcast"
	"Elevator/network/localip"
	"Elevator/network/peers"
	"Elevator/request"
	"fmt"
	"time"
)

// States of all other elevators
var GlobalElevatorStates map[string](ElevatorState)
var ConnectedNodes []string

// Only needs to be determined once on startup
var LocalID = GetID()

type ElevatorStateMessage struct {
	ID    string
	State ElevatorState
}

type ElevatorState struct {
	Floor           int
	Direction       elevio.MotorDirection
	GlobalCabOrders map[string]([config.N_FLOORS]request.RequestState)
	LocalRequests   [config.N_FLOORS][config.N_BUTTONS]request.RequestState
}

func GetCabOrdersFromNetwork() map[string]([config.N_FLOORS]request.RequestState) {
	retval := make(map[string]([config.N_FLOORS]request.RequestState))
	for id, syncState := range GlobalElevatorStates {
		var relevant [config.N_FLOORS]request.RequestState
		for floor := 0; floor < config.N_FLOORS; floor++ {
			relevant[floor] = syncState.LocalRequests[floor][elevio.BT_Cab]
		}
		retval[id] = relevant
	}
	return retval
}

// From other elevators, gets hall request states for a relevant hall button or local version of our cab requests for cab button.
func GetRequestStatesAtIndex(floor int, button elevio.ButtonType) []request.RequestState {
	var retval []request.RequestState
	var states = GlobalElevatorStates

	for _, nodeRequests := range states {
		if button == elevio.BT_Cab {
			var state = nodeRequests.GlobalCabOrders[LocalID][floor]
			retval = append(retval, state)
		} else {
			var state = nodeRequests.LocalRequests[floor][button]
			retval = append(retval, state)
		}
	}
	return retval
}

func GetID() string {
	// We assume one elevator per local ip because of hardware limits.
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	return localIP
}

func InitSyncReciever() {
	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	syncRxCh := make(chan ElevatorStateMessage)
	go peers.Receiver(config.PEER_MANAGEMENT_PORT, peerUpdateCh)
	go bcast.Receiver(config.BROADCAST_PORT, syncRxCh)
	for {
		select {
		case p := <-peerUpdateCh:
			if p.New != "" || len(p.Lost) > 0 {
				ConnectedNodes = p.Peers
			}
		case m := <-syncRxCh:
			state, ok := GlobalElevatorStates[m.ID]
			if ok {
				GlobalElevatorStates[m.ID] = state
				fmt.Println("ID: " + m.ID)
				fmt.Printf("%#v", state)
				fmt.Println("")
			}
		}
	}

}

func BroadcastState(floor *int, direction *elevio.MotorDirection, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	syncTxCh := make(chan ElevatorStateMessage)
	go bcast.Transmitter(config.BROADCAST_PORT, syncTxCh)
	for {
		var cabOrders = GetCabOrdersFromNetwork()
		syncTxCh <- ElevatorStateMessage{LocalID, ElevatorState{*floor, *direction, cabOrders, *requests}}
		time.Sleep(250 * time.Millisecond)
	}
}
