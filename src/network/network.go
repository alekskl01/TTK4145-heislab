package network

import (
	"Elevator/config"
	"Elevator/elevio"
	"Elevator/network/bcast"
	"Elevator/network/localip"
	"Elevator/network/peers"
	"Elevator/request"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// States of all other elevators
var GlobalElevatorStates sync.Map

// Potentially includes ourself
var ConnectedNodes []string

var LastRequestUpdateTime time.Time

// Only needs to be determined once on startup
var LocalID string 

type SyncMessage struct {
	ID    string
	State SyncState
}

type SyncState struct {
	Floor             int
	Direction         elevio.MotorDirection
	IsObstructed      bool
	GlobalCabOrders   map[string]([config.N_FLOORS]request.RequestState)
	RequestUpdateTime time.Time
	LocalRequests     [config.N_FLOORS][config.N_BUTTONS]request.RequestState
}

func GetOtherConnectedNodes() []string {
	var retval []string
	var allNodes = ConnectedNodes
	for _, node := range allNodes {
		if node != LocalID {
			retval = append(retval, node)
		}
	}
	return retval
}

func CheckIfNodeIsConnected(id string) bool {
	// Our local node is always connected to itself
	if id == LocalID {
		return true
	}
	var nodes = ConnectedNodes
	for _, node := range nodes {
		if node == id {
			return true
		}
	}
	return false
}

func GetCabOrdersFromNetwork() map[string]([config.N_FLOORS]request.RequestState) {
	retval := make(map[string]([config.N_FLOORS]request.RequestState))

	for _, node := range GetOtherConnectedNodes() {
		state, ok := GlobalElevatorStates.Load(node)
		if ok {
			var relevant [config.N_FLOORS]request.RequestState
			for floor := 0; floor < config.N_FLOORS; floor++ {
				relevant[floor] = state.(SyncState).LocalRequests[floor][elevio.BT_Cab]
			}
			retval[node] = relevant
		}
	}
	return retval
}

func GetLocalCabOrdersFromNetwork() []request.RequestState {
	retval := make(([]request.RequestState), config.N_FLOORS)
	for i := range retval {
		retval[i] = 0
	}
	for _, node := range GetOtherConnectedNodes() {
		state, ok1 := GlobalElevatorStates.Load(node)
		if ok1 {
			cab_orders, ok2 := state.(SyncState).GlobalCabOrders[LocalID]
			if ok2 {
				// Get the union of highest values for each request state
				for i, union_val := range retval {
					if cab_orders[i] > union_val {
						retval[i] = cab_orders[i]
					}
				}
			}
		}
	}
	return retval
}

func GetNewestOrdersFromNetwork() ([config.N_FLOORS][config.N_BUTTONS]request.RequestState, time.Time) {
	var nodes = GetOtherConnectedNodes()
	var newestTime = time.Time{}
	var newestState SyncState
	for _, node := range nodes {
		state, ok := GlobalElevatorStates.Load(node)
		if ok && (state.(SyncState)).RequestUpdateTime.After(newestTime) {
			newestState = state.(SyncState)
		}
	}
	return newestState.LocalRequests, newestTime
}

// From other elevators, gets hall request states for a relevant hall button or local version of our cab requests for cab button.
func GetRequestStatesAtIndex(floor int, button elevio.ButtonType) []request.RequestState {
	var retval []request.RequestState

	for _, node := range GetOtherConnectedNodes() {
		totalState, ok := GlobalElevatorStates.Load(node)
		if ok {
			if button == elevio.BT_Cab {
				var state = totalState.(SyncState).GlobalCabOrders[LocalID][floor]
				retval = append(retval, state)
			} else {
				var state = totalState.(SyncState).LocalRequests[floor][button]
				retval = append(retval, state)
			}
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
	
	return (localIP + ":" + strconv.Itoa(config.Port))
}

func InitSyncReciever(peerTxEnable <-chan bool, requestsUpdate chan<- [config.N_FLOORS][config.N_BUTTONS]request.RequestState, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	LastRequestUpdateTime = time.Now()
	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	go peers.Transmitter(config.PEER_MANAGEMENT_PORT, LocalID, peerTxEnable)
	peerUpdateCh := make(chan peers.PeerUpdate)
	syncRxCh := make(chan SyncMessage)
	go peers.Receiver(config.PEER_MANAGEMENT_PORT, peerUpdateCh)
	go bcast.Receiver(config.BROADCAST_PORT, syncRxCh)
	for {
		select {
		case p := <-peerUpdateCh:
			ConnectedNodes = p.Peers
			if p.New != "" {
				hallOrders, newestTime := GetNewestOrdersFromNetwork()
				if newestTime.After(LastRequestUpdateTime) {
					var newRequests = *requests
					for floor := 0; floor < config.N_FLOORS; floor++ {
						for button := 1; button < config.N_BUTTONS; button++ {
							newRequests[floor][button] = hallOrders[floor][button]
						}
					}
					requestsUpdate <- newRequests
				}
			}
		case m := <-syncRxCh:
			if m.ID != LocalID { // We are not interested in our own state
				LastRequestUpdateTime = time.Now()
				GlobalElevatorStates.Store(m.ID, m.State)
			}
		}
	}
}

func BroadcastState(floor *int, direction *elevio.MotorDirection, is_obstructed *bool, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	syncTxCh := make(chan SyncMessage)
	go bcast.Transmitter(config.BROADCAST_PORT, syncTxCh)
	for {
		var cabOrders = GetCabOrdersFromNetwork()
		syncTxCh <- SyncMessage{LocalID, SyncState{*floor, *direction, *is_obstructed, cabOrders, LastRequestUpdateTime, *requests}}
		time.Sleep(250 * time.Millisecond)
	}
}

func NetworkCheck(requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	for {
		fmt.Println("-----------------------------")
		fmt.Println("Network states:")
		PrintSyncMap(GlobalElevatorStates)
		fmt.Println("Connected nodes:")
		fmt.Printf("%#v", GetOtherConnectedNodes())
		fmt.Println()
		fmt.Println("-----------------------------")
		fmt.Println("Request matrix:")
		fmt.Printf("%#v", requests)
		fmt.Println()
		fmt.Println("-----------------------------")
		time.Sleep(5 * time.Second)
	}
}

func PrintSyncMap(m sync.Map) {
	// print map,
	fmt.Println("map content:")
	i := 0
	m.Range(func(key, value interface{}) bool {
		fmt.Printf("\t[%d] key: %v, value: %v\n", i, key, value)
		i++
		return true
	})
}
