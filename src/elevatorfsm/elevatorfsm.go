package elevatorfsm

import (
	"Elevator/elevio"
	"fmt"
)

//func elevator_setReady() //	INITIAL ELEVATOR STATE
//func getOrder()          //	CONTINUALLY CHECKS BUTTONS, STORES ORDERS
//func priorityQueue()     //	PRIORITY ALGORITHM FOR ORDER QUEUE
//func elevatorActive()    //	EXECUTE ORDERS, IDLE READY STATE IF NOT

type State int

const (
	IDLE       State = 0
	IDLE_READY       = 1
	MOVING           = 2
	DOOR_OPEN        = 3
)

type Elevator struct {
	State     State
	Floor     int
	Direction elevio.MotorDirection
	Requests  [][]int
}

func SetReady(system Elevator) {
	fmt.Println("PREPARING ELEVATOR...")
	elevio.DefaultInit()
	for {
	}
}
