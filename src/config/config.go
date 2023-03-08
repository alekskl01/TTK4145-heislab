package config

import "time"


const N_FLOORS int = 4
const N_BUTTONS int = 3
const ADDR string = "localhost:15657"

const DOOR_OPEN_DURATION = time.Millisecond * 3000
const MOTOR_STOP_DETECTION_TIME = time.Millisecond * 3000

const FSM_CHANNEL_BUFFER_SIZE = 10
