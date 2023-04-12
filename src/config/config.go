package config

import "time"

const N_FLOORS int = 4
const N_BUTTONS int = 3
const ADDR string = "localhost:12345"

// For a reasonable number of floors,
// these values should be the highest and lowest values
// returned by the hall order cost function.
const HIGH_COST int = N_FLOORS * 100
const LOW_COST int = N_FLOORS * -100
const MAJOR_COST int = N_FLOORS * 5

const DOOR_OPEN_DURATION = time.Millisecond * 3000
const MOTOR_STOP_DETECTION_TIME = time.Millisecond * 3000

const FSM_CHANNEL_BUFFER_SIZE = 10

const PEER_MANAGEMENT_PORT = 27182
const BROADCAST_PORT = 27183
