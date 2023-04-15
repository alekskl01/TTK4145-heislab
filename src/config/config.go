package config

import (
	"strconv"
	"time"
)

// Determines how often system values are refreshed and or updated.
const UPDATE_DELAY = 100 * time.Millisecond

// A time duration such that there is a high chance of a succesful update having been performed.
const SIGNIFICANT_DELAY = 4 * UPDATE_DELAY
const ELEVIO_POLL_RATE = 20 * time.Millisecond

const N_FLOORS int = 4
const N_BUTTONS int = 3

// For a reasonable number of floors,
// these values should be the highest and lowest values
// returned by the hall order cost function.
const HIGH_COST = 100 * N_FLOORS
const LOW_COST = -100 * N_FLOORS
const MAJOR_COST = 5 * N_FLOORS

const DOOR_OPEN_DURATION = 3000 * time.Millisecond
const MOTOR_STOP_DETECTION_TIME = 3000 * time.Millisecond

const FSM_CHANNEL_BUFFER_SIZE = 10

const PEER_MANAGEMENT_PORT = 27182
const BROADCAST_PORT = 27183

const DEFAULT_PORT = 15657

const IP = "localhost"

var Port = DEFAULT_PORT

// Address to connect to elevator server.
func GetAddress() string {
	return IP + ":" + strconv.Itoa(Port)
}
