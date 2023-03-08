package network

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network/bcast"
	"Elevator/network/localip"
	"Elevator/network/peers"
	"fmt"
	"os"
)

type StateMessage struct {
	senderId  string
	floor     string
	direction elevio.MotorDirection
}

type ActionType int

const (
	newRequest      ActionType = 0
	finishedRequest ActionType = 1
)

// Either represents a new request being added or a request being fulfilled.
type ActionMessage struct {
	senderId   string
	floor      string
	actionType ActionType
}

func GetID() string {
	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	return fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
}

func InitPeerManagement(id string, peerUpdateCh chan peers.PeerUpdate) {
	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	//peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(config.PEER_MANAGEMENT_PORT, id, peerTxEnable)
	go peers.Receiver(config.PEER_MANAGEMENT_PORT, peerUpdateCh)
}

func InitStateSynchronizationChannels(id string, stateTxCh chan StateMessage, stateRxCh chan StateMessage) {
	go bcast.Transmitter(config.STATE_BROADCAST_PORT, stateTxCh)
	go bcast.Receiver(config.STATE_BROADCAST_PORT, stateRxCh)
}

func InitActionSynchronizationChannels(id string, actionTxCh chan ActionMessage, actionRxCh chan ActionMessage) {
	go bcast.Transmitter(config.STATE_BROADCAST_PORT, actionTxCh)
	go bcast.Receiver(config.STATE_BROADCAST_PORT, actionRxCh)
}
