// This package contains all functionality that involves network communication with others nodes.
package network

import (
	"Elevator/config"
	"Elevator/cost"
	"Elevator/elevatorstate"
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
var _globalElevatorStates sync.Map

// Newest Peer list from peer synchronization, may include ourself
var ConnectedNodes []string

// Last time this node recieved a state sync message from another node
// and was not in the process of resynchronizing with the network (see isSynchronized).
var _lastRequestUpdateTime = time.Time{}

// Used to intermittently disable use of local state info for actions during resynchronization with network.
var _isSynchronized = false

// Only needs to be determined once on startup,
// consists of the network local ip of this machine and the port being used for
// elevator communication on startup.
var LocalID string

type SyncMessage struct {
	ID    string
	State SyncState
}

var _globalCabOrders map[string]([config.N_FLOORS]request.RequestState)

type SyncState struct {
	// Local elevator data
	Floor     int
	Direction elevio.MotorDirection
	// Tells us at what point in the state machine the elevator is in
	FSMState      elevatorstate.ElevatorState
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

func GetID() string {
	// We assume one elevator per local ip because of hardware limits.
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}

	return (localIP + ":" + strconv.Itoa(config.Port))
}

// For most cases it is not interesting to know that this node is connected to itself.
func getOtherConnectedNodes() []string {
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
	// It is expected that this node is connected to itself.
	if id == LocalID {
		return true
	}
	var nodes = getOtherConnectedNodes()
	for _, node := range nodes {
		if node == id {
			return true
		}
	}
	return false
}

// Checks all connected nodes to see if it is 'cheapest' for this node to
// fulfill a specific hall order. Weighting based on the cost function.
// Important note: returns true if our cost is the same as the cost for other nodes,
// this is vital to ensure that orders are handled.
func IsHallOrderCheapest(hallFloor int, buttonType elevio.ButtonType, floor *int, direction *elevio.MotorDirection,
	fsm_state *elevatorstate.ElevatorState, isObstructed *bool, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) bool {
	// Noone else to take it
	if len(getOtherConnectedNodes()) == 0 {
		return true
	}
	cheapestCost := config.HIGH_COST
	for _, node := range getOtherConnectedNodes() {
		stateAsAny, ok := _globalElevatorStates.Load(node)
		if ok {
			state := stateAsAny.(SyncState)
			cost := cost.GetCostOfHallOrder(hallFloor, buttonType, state.Floor, state.Direction, state.FSMState, state.IsObstructed, state.LocalRequests)
			if cost < cheapestCost {
				cheapestCost = cost
			}
		}
	}

	ourCost := cost.GetCostOfHallOrder(hallFloor, buttonType, *floor, *direction, *fsm_state, *isObstructed, *requests)
	return ourCost <= cheapestCost
}

// Used get the cab orders of other nodes and store them,
// so they can later be fetched with GetLocalCabOrdersFromNetwork() in case they go down.
// Trusts that the newest cab orders given by a node are correct.
func getCabOrdersFromNetwork() {
	for _, node := range getOtherConnectedNodes() {
		state, ok := _globalElevatorStates.Load(node)
		if ok {
			var nodeCabOrders [config.N_FLOORS]request.RequestState
			for floor := 0; floor < config.N_FLOORS; floor++ {
				nodeCabOrders[floor] = state.(SyncState).LocalRequests[floor][elevio.BT_Cab]
			}
			_globalCabOrders[node] = nodeCabOrders
		}
	}
}

// Used to fetch this node's cab orders from before going down stored by other nodes.
// These values are stored using GetLocalCabOrdersFromNetwork().
// To ensure no orders are lost we take the highest RequestState union of those stored by all nodes.
// (this may lead to some excess fulfillment in some cases)
func GetLocalCabOrdersFromNetwork() []request.RequestState {
	cabOrderUnion := make(([]request.RequestState), config.N_FLOORS)
	for i := range cabOrderUnion {
		cabOrderUnion[i] = 0
	}
	for _, node := range getOtherConnectedNodes() {
		state, ok1 := _globalElevatorStates.Load(node)
		if ok1 {
			cabOrders, ok2 := state.(SyncState).GlobalCabOrders[LocalID]
			if ok2 {
				// Get the union of highest values for each request state
				for i, unionState := range cabOrderUnion {
					if cabOrders[i] > unionState {
						cabOrderUnion[i] = cabOrders[i]
					}
				}
			}
		}
	}
	return cabOrderUnion
}

// Used to synchronize with network, at startup and when connection to the network is lost.
// In such a case the network is standardized to use the newest synchronized state.
// Note: it is important to iterate through the list of nodes in a consistently sorted order,
// including this node to avoid errors when multiple nodes have the same update time.
func getNewestRequestsFromNetwork() ([config.N_FLOORS][config.N_BUTTONS]request.RequestState, bool, bool) {
	var nodes = getOtherConnectedNodes()
	nodes = append(nodes, LocalID)
	sort.Strings(nodes)
	var newestTime = time.Time{}
	var newestState SyncState
	// Is the local state the newest one?
	var useLocalState = false
	// Do we only have access to the local state?
	var onlyLocalState = true
	for _, node := range nodes {
		if node == LocalID {
			if _lastRequestUpdateTime.After(newestTime) {
				useLocalState = true
				newestTime = _lastRequestUpdateTime
			}
		} else {
			state, ok := _globalElevatorStates.Load(node)
			if ok && (state.(SyncState)).RequestUpdateTime.After(newestTime) {
				useLocalState = false
				onlyLocalState = false
				newestState = state.(SyncState)
				newestTime = newestState.RequestUpdateTime
			}
		}
	}
	return newestState.LocalRequests, useLocalState, onlyLocalState
}

// From other elevators, gets hall request states for a relevant hall button or their local version of our cab requests for cab buttons.
// Note that cab orders are ensured to be stored by other nodes before locally being set to active.
func GetRequestStatesAtIndex(floor int, button elevio.ButtonType) ([]request.RequestState, bool) {
	var indexStates []request.RequestState
	// Important to avoid
	var anyNotSynchronized = false
	for _, node := range getOtherConnectedNodes() {
		stateAsAny, ok := _globalElevatorStates.Load(node)
		if ok {
			var state = stateAsAny.(SyncState)
			// Ignore states from nodes mid way in synchronization
			if state.IsSynchronized {
				if button == elevio.BT_Cab {
					var state = stateAsAny.(SyncState).GlobalCabOrders[LocalID][floor]
					indexStates = append(indexStates, state)
				} else {
					var state = stateAsAny.(SyncState).LocalRequests[floor][button]
					indexStates = append(indexStates, state)
				}
			} else {
				anyNotSynchronized = true
			}
		}
	}
	return indexStates, anyNotSynchronized
}

// Synchronizes this node's hall orders with the newest ones on the network,
// allowing for this node take part in regular state synchronization.
func delayedResynchronization(requestsUpdateCh chan<- [config.N_FLOORS][config.N_BUTTONS]request.RequestState, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	// Ensure we have enough time to get updated states from network
	time.Sleep(config.SIGNIFICANT_DELAY)
	if !_isSynchronized {
		log("Attempting resynchronization")
		hallOrders, useLocalState, onlyLocalState := getNewestRequestsFromNetwork()
		// Ensure that we have had the opportunity to get at least 1 state from another node.
		if !onlyLocalState {
			if !useLocalState {
				log("Synching to external state")
				var newRequests = *requests
				for floor := 0; floor < config.N_FLOORS; floor++ {
					for button := 0; button < (config.N_BUTTONS - 1); button++ {
						newRequests[floor][button] = hallOrders[floor][button]
					}
				}
				requestsUpdateCh <- newRequests
			} else {
				log("Synching to own state")
			}
			// We have resynchronized with the network, enable regular state synchronization.
			log("Reconnected and resynchronized, useLocalState?  " + strconv.FormatBool(useLocalState))
			_isSynchronized = true
		}
	}
}

// Handles peer node management, when other nodes and connected or disconnected to or from this one.
// Note that specific resynchronization handling is needed when a new node is connected and this node has been disconnected.
func PeerUpdateReciever(peerTxEnableCh <-chan bool, requestsUpdateCh chan<- [config.N_FLOORS][config.N_BUTTONS]request.RequestState, requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	go peers.Transmitter(config.PEER_MANAGEMENT_PORT, LocalID, peerTxEnableCh)
	peerUpdateCh := make(chan peers.PeerUpdate)
	go peers.Receiver(config.PEER_MANAGEMENT_PORT, peerUpdateCh)
	for {
		select {
		case p := <-peerUpdateCh:
			ConnectedNodes = p.Peers
			if len(getOtherConnectedNodes()) == 0 {
				// We are disconnected from the newtwork, disable broadcast.
				_isSynchronized = false
				_globalElevatorStates = sync.Map{}
				log("Disconnected from network")
			}
			if p.New != "" && p.New != LocalID {
				log("New peer detected, resynchronizing.")
				go delayedResynchronization(requestsUpdateCh, requests)
			}
			// This should come up very rarely, but serves to ensure that this node tries to resynchronize.
			if (p.New == "" || p.New == LocalID) && !_isSynchronized {
				go delayedResynchronization(requestsUpdateCh, requests)
			}
		}
	}
}

func SyncReciever() {
	syncRxCh := make(chan SyncMessage)
	go bcast.Receiver(config.BROADCAST_PORT, syncRxCh)
	for {
		select {
		case m := <-syncRxCh:
			if m.ID != LocalID { // We are not interested in our own state
				if _isSynchronized {
					// This node is not considered updated if it has not resynchronized with the network.
					_lastRequestUpdateTime = time.Now()
				}
				_globalElevatorStates.Store(m.ID, m.State)
			}
		}
	}
}

func BroadcastState(floor *int, direction *elevio.MotorDirection, state *elevatorstate.ElevatorState, is_obstructed *bool,
	requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	_globalCabOrders = make(map[string]([config.N_FLOORS]request.RequestState))
	syncTxCh := make(chan SyncMessage)
	go bcast.Transmitter(config.BROADCAST_PORT, syncTxCh)
	for {
		// Update global cab orders before broadcasting.
		getCabOrdersFromNetwork()
		syncTxCh <- SyncMessage{LocalID, SyncState{*floor, *direction, *state, *is_obstructed,
			*requests, _isSynchronized, _globalCabOrders, _lastRequestUpdateTime}}
		time.Sleep(config.UPDATE_DELAY)
	}
}

// Used to debug network failures.
func NetworkCheck(requests *[config.N_FLOORS][config.N_BUTTONS]request.RequestState) {
	for {
		fmt.Println("-----------------------------")
		fmt.Println("Network states:")
		PrintSyncMap(_globalElevatorStates)
		fmt.Println("Connected nodes:")
		fmt.Printf("%#v", getOtherConnectedNodes())
		fmt.Println()
		fmt.Println("Global Cab Orders:")
		fmt.Printf("%#v", _globalCabOrders)
		fmt.Println()
		fmt.Println("-----------------------------")
		fmt.Println("Request matrix:")
		fmt.Printf("%#v", requests)
		fmt.Println()
		fmt.Println("-----------------------------")
		time.Sleep(5 * time.Second)
	}
}

// Copied from https://stackoverflow.com/questions/58995416/how-to-pretty-print-the-contents-of-a-sync-map
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
