package config

import (
	"strconv"
	"time"
)

// Determines how often system values are refreshed and or updated.
const UPDATE_DELAY = time.Millisecond * 250

const N_FLOORS int = 4
const N_BUTTONS int = 3

// For a reasonable number of floors,
// these values should be the highest and lowest values
// returned by the hall order cost function.
const HIGH_COST = N_FLOORS * 100
const LOW_COST = N_FLOORS * -100
const MAJOR_COST = N_FLOORS * 5

const DOOR_OPEN_DURATION = time.Millisecond * 3000
const MOTOR_STOP_DETECTION_TIME = time.Millisecond * 3000

const FSM_CHANNEL_BUFFER_SIZE = 10

const PEER_MANAGEMENT_PORT = 27182
const BROADCAST_PORT = 27183

const DEFAULT_PORT = 15657

const IP = "localhost"
var Port = DEFAULT_PORT

func GetAddress() string {
	return IP + ":" + strconv.Itoa(Port)
}
