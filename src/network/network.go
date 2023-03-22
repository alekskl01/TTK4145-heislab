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

var networkRequests map[string](SyncState)
var connectedNodes []string

// Only needs to be determined once on startup
var localID = GetID()

type SyncMessage struct {
	ID    string
	State SyncState
}

type SyncState struct {
	Floor     int
	Direction elevio.MotorDirection
	Requests  [config.N_FLOORS][config.N_BUTTONS]request.RequestState
}

func GetRequestStatesAtIndex(floor int, button elevio.ButtonType) []request.RequestState {
	var retval []request.RequestState
	var requests = networkRequests

	for _, nodeRequests := range requests {
		var state = nodeRequests.Requests[floor][button]
		retval = append(retval, state)
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
	syncRxCh := make(chan SyncMessage)
	go peers.Receiver(config.PEER_MANAGEMENT_PORT, peerUpdateCh)
	go bcast.Receiver(config.BROADCAST_PORT, syncRxCh)
	for {
		select {
		case p := <-peerUpdateCh:
			if p.New != "" || len(p.Lost) > 0 {
				connectedNodes = p.Peers
			}
		case m := <-syncRxCh:
			state, ok := networkRequests[m.ID]
			if !ok || state != m.State {
				networkRequests[m.ID] = state
			}
		}
	}
}

func BroadcastState(floor *int, direction *elevio.MotorDirection, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	syncTxCh := make(chan SyncMessage)
	go bcast.Transmitter(config.BROADCAST_PORT, syncTxCh)
	for {
		syncTxCh <- SyncMessage{GetID(), SyncState{*floor, *direction, *requests}}
		time.Sleep(250 * time.Millisecond)
	}
}
