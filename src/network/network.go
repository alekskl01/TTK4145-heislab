package network

import (
	"Elevator/config"
	"Elevator/cost"
	"Elevator/elevio"
	"Elevator/network/bcast"
	"Elevator/network/localip"
	"Elevator/network/peers"
	"Elevator/request"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"
)

// States of all other elevators
var GlobalElevatorStates sync.Map

// Potentially includes ourself
var ConnectedNodes []string

var LastRequestUpdateTime time.Time

// Used to intermittently disable use of local state info for actions during resynchronization with network.
var isSynchronized = true

// Only needs to be determined once on startup
var LocalID string

type SyncMessage struct {
	ID    string
	State SyncState
}

type SyncState struct {
	// Local elevator data
	Floor         int
	Direction     elevio.MotorDirection
	IsObstructed  bool
	LocalRequests [config.N_FLOORS][config.N_BUTTONS]request.RequestState

	// Metadata for node synchronization
	IsSynchronized    bool
	GlobalCabOrders   map[string]([config.N_FLOORS]request.RequestState)
	RequestUpdateTime time.Time
}

func log(text string) {
	fmt.Println("Network: " + text)
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

func IsHallOrderCheapest(hall_floor int, button_type elevio.ButtonType, floor *int, direction *elevio.MotorDirection,
	is_obstructed *bool, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) bool {
	if len(GetOtherConnectedNodes()) == 0 {
		return true
	}
	cheapest_cost := config.HIGH_COST
	for _, node := range GetOtherConnectedNodes() {
		state, ok := GlobalElevatorStates.Load(node)
		if ok {
			sync_state := state.(SyncState)
			cost := cost.GetCostOfHallOrder(hall_floor, button_type, sync_state.Floor, sync_state.Direction, sync_state.IsObstructed, sync_state.LocalRequests)
			if cost < cheapest_cost {
				cheapest_cost = cost
			}
		}
	}

	our_cost := cost.GetCostOfHallOrder(hall_floor, button_type, *floor, *direction, *is_obstructed, *requests)
	return our_cost <= cheapest_cost
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

func GetNewestOrdersFromNetwork() ([config.N_FLOORS][config.N_BUTTONS]request.RequestState, bool) {
	var nodes = GetOtherConnectedNodes()
	nodes = append(nodes, LocalID)
	sort.Strings(nodes)
	var newestTime = time.Time{}
	var newestState SyncState
	var useLocalState = false
	for _, node := range nodes {
		if node == LocalID {
			if LastRequestUpdateTime.After(newestTime) {
				useLocalState = true
				newestTime = LastRequestUpdateTime
			}
		} else {
			state, ok := GlobalElevatorStates.Load(node)
			if ok && (state.(SyncState)).RequestUpdateTime.After(newestTime) {
				useLocalState = false
				newestState = state.(SyncState)
				newestTime = newestState.RequestUpdateTime
			}
		}
	}
	return newestState.LocalRequests, useLocalState
}

// From other elevators, gets hall request states for a relevant hall button or local version of our cab requests for cab button.
func GetRequestStatesAtIndex(floor int, button elevio.ButtonType) []request.RequestState {
	var retval []request.RequestState

	for _, node := range GetOtherConnectedNodes() {
		stateAsAny, ok := GlobalElevatorStates.Load(node)
		if ok {
			var state = stateAsAny.(SyncState)
			// Ignore states from nodes mid way in synchronization
			if state.IsSynchronized {
				if button == elevio.BT_Cab {
					var state = stateAsAny.(SyncState).GlobalCabOrders[LocalID][floor]
					retval = append(retval, state)
				} else {
					var state = stateAsAny.(SyncState).LocalRequests[floor][button]
					retval = append(retval, state)
				}
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

func DelayedResynchronization(requestsUpdate chan<- [config.N_FLOORS][config.N_BUTTONS]request.RequestState, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	// Ensure we have enough time to get updated states from network
	time.Sleep(4 * config.UPDATE_DELAY)
	if !isSynchronized {
		hallOrders, useLocalState := GetNewestOrdersFromNetwork()
		if !useLocalState {
			var newRequests = *requests
			for floor := 0; floor < config.N_FLOORS; floor++ {
				for button := 1; button < config.N_BUTTONS; button++ {
					newRequests[floor][button] = hallOrders[floor][button]
				}
			}
			requestsUpdate <- newRequests
		}
		// We have resynchronized with the network, enable broadcast.
		log("Reconnected and resynchronized, useLocalState?  " + strconv.FormatBool(useLocalState))
		isSynchronized = true
	}
}

func PeerUpdateReciever(peerTxEnable <-chan bool, requestsUpdate chan<- [config.N_FLOORS][config.N_BUTTONS]request.RequestState, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	go peers.Transmitter(config.PEER_MANAGEMENT_PORT, LocalID, peerTxEnable)
	peerUpdateCh := make(chan peers.PeerUpdate)
	go peers.Receiver(config.PEER_MANAGEMENT_PORT, peerUpdateCh)
	for {
		select {
		case p := <-peerUpdateCh:
			if len(p.Peers) == 0 {
				// We are disconnected from the newtwork, disable broadcast.
				isSynchronized = false
				log("Disconnected from network")
			}

			ConnectedNodes = p.Peers
			if p.New != "" {
				go DelayedResynchronization(requestsUpdate, requests)
			}
		}
	}
}

func SyncReciever() {
	LastRequestUpdateTime = time.Now()
	syncRxCh := make(chan SyncMessage)
	go bcast.Receiver(config.BROADCAST_PORT, syncRxCh)
	for {
		select {
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
		syncTxCh <- SyncMessage{LocalID, SyncState{*floor, *direction, *is_obstructed, *requests, isSynchronized, cabOrders, LastRequestUpdateTime}}
		time.Sleep(config.UPDATE_DELAY)
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
